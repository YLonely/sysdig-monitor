package container

import (
	"context"

	"github.com/YLonely/sysdig-monitor/log"
	"github.com/YLonely/sysdig-monitor/server/model"
	"github.com/YLonely/sysdig-monitor/sysdig"
)

func processLoop(ctx context.Context, c *model.Container, ch chan sysdig.Event) {
	log.L.WithField("container-id", c.ID).Info("process loop start.")
	for e := range ch {

	}
	log.L.WithField("container-id", c.ID).Error("process loop unexpected exited.")
}
