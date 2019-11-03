package container

import (
	"github.com/YLonely/sysdig-monitor/server/controller"
	"github.com/YLonely/sysdig-monitor/server/model"
	"github.com/YLonely/sysdig-monitor/server/router"
	"github.com/YLonely/sysdig-monitor/sysdig"
)

// Container represents a top level controller for container
type ContainerController struct {
	containerRouter router.Router
	// event chan
	ec chan sysdig.Event

	// containers
	containers map[string]*model.Container
	// container process channels
	containerCh map[string]chan sysdig.Event
}

func NewController(ec chan sysdig.Event) controller.Controller {
	r := router.NewGroupRouter("/container")
	res := &ContainerController{containerRouter: r, ec: ec, containers: map[string]*model.Container{}}
	res.initRouter()
	return res
}

var _ controller.Controller = &ContainerController{}

func (cc *ContainerController) BindedRoutes() []router.Route {
	return cc.containerRouter.Routes()
}

func (cc *ContainerController) initRouter() {
	// TODO
}
