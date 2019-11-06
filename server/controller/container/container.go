package container

import (
	"context"
	"sync"

	"github.com/YLonely/sysdig-monitor/log"
	"github.com/YLonely/sysdig-monitor/server/controller"
	"github.com/YLonely/sysdig-monitor/server/model"
	"github.com/YLonely/sysdig-monitor/server/router"
	"github.com/YLonely/sysdig-monitor/sysdig"
	"github.com/docker/docker/client"
    "github.com/docker/docker/api/types"

)

// Container represents a top level controller for container
type ContainerController struct {
	containerRouter router.Router
	// event chan
	ec chan sysdig.Event

	// containers use container id as key
	containers map[string]*container
	// container process channels
	containerCh map[string]chan containerEvent

	//containers mutex
	cm sync.RWMutex

	//docker client
	dockerCli *client.Client
}

type container struct {
	imageConfig types.ImageInspect
	m sync.RWMutex
	*model.Container
}

func newContainer(id, name string) *container {
	c := model.NewContainer(id, name)
	return &container{Container: c}
}

const eventBufferLen = 512
const unknownContainerName = "<unknown>"
const incompleteContainerName = "incomplete"

func NewController(ctx context.Context, serverErrorChannel chan<- error) (controller.Controller, error) {
	r := router.NewGroupRouter("/container")
	sysdigServer := sysdig.NewServer()
	res := &ContainerController{containerRouter: r, ec: sysdigServer.Subscribe(), containers: map[string]*container{}, containerCh: map[string]chan containerEvent{}}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	res.dockerCli = cli

	if err := res.start(ctx); err != nil {
		return nil, err
	}
	sysdigServC, err := sysdigServer.Start(ctx)
	if err != nil {
		return nil, err
	}
	go func() {
		e := <-sysdigServC
		serverErrorChannel <- e
	}()
	res.initRouter()
	return res, nil
}

var _ controller.Controller = &ContainerController{}

func (cc *ContainerController) BindedRoutes() []router.Route {
	return cc.containerRouter.Routes()
}

func (cc *ContainerController) initRouter() {
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
				container := newcontainer(ce.containerID, containerName)
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
