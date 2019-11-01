package controller

import (
	"github.com/YLonely/sysdig-monitor/server/router"
)

type Controller interface {
	BindedRoutes() []router.Route
}
