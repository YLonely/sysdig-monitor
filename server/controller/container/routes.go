package container

import (
	"github.com/YLonely/sysdig-monitor/server/model"
	"github.com/gin-gonic/gin"
)

func (cc *ContainerController) getAllContainers(c *gin.Context) {
	res := map[string]string{}
	cc.cm.RLock()
	for id, container := range cc.containers {
		res[id] = container.Name
	}
	cc.cm.RUnlock()
	c.JSON(200, res)
}

type FlattenConnection struct {
	model.ConnectionMeta
	model.Connection
}

type GetContainerResponse struct {
	*model.Container
	ActiveConnections []FlattenConnection `json:"active_connections"`
}

func (cc *ContainerController) getContainer(c *gin.Context) {
	cid := c.Param("id")
	cc.cm.RLock()
	container, exists := cc.containers[cid]
	cc.cm.RUnlock()
	if !exists {
		c.JSON(200, "no such container error")
		return
	}
	container.m.RLock()
	defer container.m.RUnlock()
	flattenConns := []FlattenConnection{}
	for meta, conn := range container.ActiveConnections {
		flattenConns = append(flattenConns, FlattenConnection{Connection: *conn, ConnectionMeta: meta})
	}
	c.JSON(200, GetContainerResponse{Container: container.Container, ActiveConnections: flattenConns})
}
