package controller

import (
	"github.com/YLonely/sysdig-monitor/server/router"
)

type Controller interface {
	// BindedRoutes return all the routes binded to the controller
	BindedRoutes() []router.Route
	// Release releases all the occupied resources which should be handled properly
	Release()
}
