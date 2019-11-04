package container

import (
	"context"
	"sync"

	"github.com/YLonely/sysdig-monitor/log"
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

	// containers use container id as key
	containers map[string]*model.Container
	// container process channels
	containerCh map[string]chan containerEvent

	//containers mutex
	cm sync.RWMutex
}

const eventBufferLen = 512

func NewController(ctx context.Context, ec chan sysdig.Event) (controller.Controller, error) {
	r := router.NewGroupRouter("/container")
	res := &ContainerController{containerRouter: r, ec: ec, containers: map[string]*model.Container{}, containerCh: map[string]chan containerEvent{}}
	res.initRouter()
	if err := res.start(ctx); err != nil {
		return res, err
	}
	return res, nil
}

var _ controller.Controller = &ContainerController{}

func (cc *ContainerController) BindedRoutes() []router.Route {
	return cc.containerRouter.Routes()
}

func (cc *ContainerController) initRouter() {
	// TODO
}

func (cc *ContainerController) start(ctx context.Context) error {
	func() {
		var e sysdig.Event
		for {
			select {
			case e = <-cc.ec:
			case <-ctx.Done():
				return
			}
			if e.ContainerName == "host" {
				continue
			}
			ce := convert(e)
			containerID := ce.containerID
			if _, exists := cc.containers[containerID]; !exists {
				container := model.NewContainer(ce.containerID, ce.containerName)
				ch := make(chan containerEvent, eventBufferLen)
				cc.cm.Lock()
				cc.containers[containerID] = container
				cc.cm.Unlock()
				cc.containerCh[containerID] = ch
				go func() {
					log.L.WithField("container-id", containerID).Info("processLoop start")
					err := processLoop(ctx, container, ch)
					log.L.WithField("container-id", containerID).Info("processLoop exits")
					if err != nil {
						log.L.WithError(err).Error("")
					}
					cc.cm.Lock()
					cc.containers[container.ID] = nil
					cc.cm.Unlock()
					cc.containerCh[container.ID] = nil
					close(ch)
				}()
			}
			ch := cc.containerCh[containerID]
			ch <- ce
		}
	}()
	return nil
}

func convert(e sysdig.Event) containerEvent {
	res := containerEvent{}
	res.containerID = e.ContainerID
	res.containerName = e.ContainerName
	res.bufferLen = e.EventBuflen
	res.eventDir = e.EventDir
	res.eventType = e.EventType
	res.fdName = e.FdName
	res.fdType = e.FdType
	res.isIORead = e.EventIsIORead
	res.isIOWrite = e.EventIsIOWrite
	res.latency = e.EventLatency
	res.rawRes = e.RawRes
	res.syscallType = e.SyscallType
	res.virtualtid = e.ThreadVirtualID
	return res
}
