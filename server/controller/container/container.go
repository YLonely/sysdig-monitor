package container

import (
	"context"
	"sync"
	"time"

	"github.com/YLonely/sysdig-monitor/log"
	"github.com/YLonely/sysdig-monitor/server/controller"
	"github.com/YLonely/sysdig-monitor/server/model"
	"github.com/YLonely/sysdig-monitor/server/router"
	"github.com/YLonely/sysdig-monitor/sysdig"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type containerController struct {
	containerRouter router.Router
	// event chan
	ec chan sysdig.Event

	// containers use container id as key
	containers map[string]*mutexContainer
	// container process channels
	containerCh map[string]chan containerEvent

	//containers mutex
	cm sync.RWMutex

	//docker client
	dockerCli *client.Client

	sysdigServer sysdig.Server
}

type mutexContainer struct {
	m sync.RWMutex
	*model.Container
}

const containerKeepingPeriod = time.Millisecond * 50

func newMutexContainer(id, name string, containerJSON *types.ContainerJSON) (*mutexContainer, error) {
	c, err := model.NewContainer(id, name, containerJSON.GraphDriver.Data)
	if err != nil {
		return nil, err
	}
	return &mutexContainer{Container: c}, nil
}

const eventBufferLen = 512
const unknownContainerName = "<unknown>"
const incompleteContainerName = "incomplete"

func NewController(ctx context.Context, serverErrorChannel chan<- error) (controller.Controller, error) {
	r := router.NewGroupRouter("/container")
	sysdigServer := sysdig.NewServer(ctx)
	res := &containerController{containerRouter: r, ec: sysdigServer.Subscribe(), containers: map[string]*mutexContainer{}, containerCh: map[string]chan containerEvent{}}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	res.dockerCli = cli

	if err := res.start(ctx); err != nil {
		return nil, err
	}
	sysdigServC, err := sysdigServer.Start()
	if err != nil {
		return nil, err
	}
	res.sysdigServer = sysdigServer
	go func() {
		e := <-sysdigServC
		serverErrorChannel <- e
	}()
	res.initRouter()
	return res, nil
}

var _ controller.Controller = &containerController{}

func (cc *containerController) BindedRoutes() []router.Route {
	return cc.containerRouter.Routes()
}

func (cc *containerController) Release() {
	cc.sysdigServer.Shutdown()
}

func (cc *containerController) initRouter() {
	cc.containerRouter.AddRoute("/", router.MethodGet, cc.getAllContainers)
	cc.containerRouter.AddRoute("/:id", router.MethodGet, cc.getContainer)
}

func (cc *containerController) start(ctx context.Context) error {
	go func() {
		var (
			e      sysdig.Event
			ch     chan containerEvent
			exists bool
		)
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
			cc.cm.RLock()
			ch, exists = cc.containerCh[containerID]
			if exists {
				ch <- ce
			}
			cc.cm.RUnlock()
			if !exists {
				containerJSON, err := cc.containerJSON(ctx, containerID)
				if err != nil {
					log.L.WithError(err).WithField("container-id", containerID).Error("cant fetch container json")
					continue
				}
				container, err := newMutexContainer(ce.containerID, containerName, containerJSON)
				if err != nil {
					log.L.WithError(err).WithField("container-id", containerID).Error("")
					continue
				}
				ch := make(chan containerEvent, eventBufferLen)
				cc.cm.Lock()
				cc.containers[containerID] = container
				cc.containerCh[containerID] = ch
				cc.cm.Unlock()
				go cc.containerProcessLoop(ctx, container, ch)
			}
		}
	}()
	return nil
}

func (cc *containerController) containerJSON(ctx context.Context, id string) (*types.ContainerJSON, error) {
	j, err := cc.dockerCli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (cc *containerController) containerProcessLoop(ctx context.Context, container *mutexContainer, ch chan containerEvent) {
	log.L.WithField("container-id", container.ID).Info("processLoop start")
	err := processLoop(ctx, container, ch)
	log.L.WithField("container-id", container.ID).Info("processLoop exits")
	if err != nil {
		log.L.WithError(err).Error("")
	}
	// In some cases, a few events will be catched after the exit of the container
	// so wait a period of time before clean up to prevent the creation of processLoop of the same container
	time.Sleep(containerKeepingPeriod)
	cc.cm.Lock()
	delete(cc.containers, container.ID)
	delete(cc.containerCh, container.ID)
	cc.cm.Unlock()
	close(ch)
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
