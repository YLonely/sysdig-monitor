package prometheus

import (
	"github.com/YLonely/sysdig-monitor/localprom"
	"github.com/gin-gonic/gin"
)

func (pc *prometheusContorller) metrics(c *gin.Context) {
	localprom.RunMetrics()
	pc.handler.ServeHTTP(c.Writer, c.Request)
}
