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
	containers map[string]*mutexContainer
	// container process channels
	containerCh map[string]chan containerEvent

	//containers mutex
	cm sync.RWMutex
}

type mutexContainer struct {
	m sync.RWMutex
	*model.Container
}

func newMutexContainer(id, name string) *mutexContainer {
	c := model.NewContainer(id, name)
	return &mutexContainer{Container: c}
}

const eventBufferLen = 1024
const unknownContainerName = "<unknown>"
const incompleteContainerName = "incomplete"

func NewController(ctx context.Context, ec chan sysdig.Event) (controller.Controller, error) {
	r := router.NewGroupRouter("/container")
	res := &ContainerController{containerRouter: r, ec: ec, containers: map[string]*mutexContainer{}, containerCh: map[string]chan containerEvent{}}
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
	cc.containerRouter.AddRoute("/", router.MethodGet, cc.getAllContainers)
	cc.containerRouter.AddRoute("/:id", router.MethodGet, cc.getContainer)
}

func (cc *ContainerController) start(ctx context.Context) error {
	go func() {
		var e sysdig.Event
		for {
			select {
			case e = <-cc.ec:
			case <-ctx.Done():
				return
			}
			if e.ContainerID == "host" || len(e.ContainerID) == 0 {
				continue
			}
			ce := convert(e)
			containerID := ce.containerID
			containerName := ce.containerName
			if containerName == incompleteContainerName {
				containerName = unknownContainerName
			}
			log.L.Debug(ce)
			if _, exists := cc.containers[containerID]; !exists {
				container := newMutexContainer(ce.containerID, containerName)
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
					delete(cc.containers, container.ID)
					cc.cm.Unlock()
					delete(cc.containerCh, container.ID)
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
