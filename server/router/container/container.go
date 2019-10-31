package container

import (
	"github.com/YLonely/sysdig-monitor/server/router"
)

type containerRouter struct {
	routes []router.Route
	// sysdig client

	// docker client
}

func NewRouter() router.Router {
	r := &containerRouter{}

	return r
}

func (r *containerRouter) Routes() []router.Route {
	return r.routes
}

func (r *containerRouter) initRoutes() {
	r.routes = []router.Route{}
}
