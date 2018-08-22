package models

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"time"

	"github.com/btcsuite/golangcrypto/bcrypt"
	"github.com/go-pg/pg"
	"github.com/godiscourse/godiscourse/session"
	"github.com/godiscourse/godiscourse/uuid"
)

const sessions_DDL = `
CREATE TABLE IF NOT EXISTS sessions (
	session_id            VARCHAR(36) PRIMARY KEY,
	user_id               VARCHAR(36) NOT NULL,
	secret                VARCHAR(1024) NOT NULL,
	created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX ON sessions (user_id);
CREATE INDEX ON sessions (created_at);
`

type Session struct {
	SessionId string    `sql:"session_id,pk"`
	UserId    string    `sql:"user_id"`
	Secret    string    `sql:"secret"`
	CreatedAt time.Time `sql:"created_at"`
}

var sessionCols = []string{"session_id", "user_id", "secret", "created_at"}

func CreateSession(ctx context.Context, identity, password, sessionSecret string) (*User, error) {
	data, err := hex.DecodeString(sessionSecret)
	if err != nil {
		return nil, session.BadDataError(ctx)
	}
	public, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, session.BadDataError(ctx)
	}
	switch public.(type) {
	case *ecdsa.PublicKey:
	default:
		return nil, session.BadDataError(ctx)
	}

	user, err := FindUserByUsernameOrEmail(ctx, identity)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, session.IdentityNonExistError(ctx)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(password)); err != nil {
		return nil, session.InvalidPasswordError(ctx)
	}

	err = session.Database(ctx).RunInTransaction(func(tx *pg.Tx) error {
		sess, err := user.addSession(ctx, tx, sessionSecret)
		if err != nil {
			return err
		}
		user.SessionId = sess.SessionId
		return nil
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return user, nil
}

func (user *User) addSession(ctx context.Context, tx *pg.Tx, secret string) (*Session, error) {
	sess := &Session{
		SessionId: uuid.NewV4().String(),
		UserId:    user.UserId,
		Secret:    secret,
		CreatedAt: time.Now(),
	}
	err := tx.Insert(sess)
	if err != nil {
		return nil, err
	}
	return sess, nil
}
