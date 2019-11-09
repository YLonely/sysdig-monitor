package prometheus

import (
	"github.com/gin-gonic/gin"
)

func (pc *prometheusContorller) metrics(c *gin.Context) {
	pc.handler.ServeHTTP(c.Writer, c.Request)
}
