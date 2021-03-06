package admin

import "github.com/dimfeld/httptreemux"

// RegisterAdminRoutes register admin routes
func RegisterAdminRoutes(router *httptreemux.TreeMux) {
	registerAdminUser(router)
	registerAdminCategory(router)
	registerAdminTopic(router)
}
