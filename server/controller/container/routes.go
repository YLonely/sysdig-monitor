package container

import (
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
	c.JSON(200, container.Container)
}
