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
	containers map[string]model.Container
}

func NewController(ec chan sysdig.Event) controller.Controller {
	r := router.NewGroupRouter("/container")

}

func (cc *ContainerController) initRouter() {

}
