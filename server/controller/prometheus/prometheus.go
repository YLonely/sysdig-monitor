package prometheus

import (
	"context"
	"net/http"

	"github.com/YLonely/sysdig-monitor/server/controller"
	"github.com/YLonely/sysdig-monitor/server/router"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type prometheusContorller struct {
	router  router.Router
	handler http.Handler
}

func NewController(ctx context.Context) controller.Controller {
	res := &prometheusContorller{router: router.NewRouter(), handler: promhttp.Handler()}
	res.initRouter()
	return res
}

func (pc *prometheusContorller) BindedRoutes() []router.Route {
	return pc.router.Routes()
}

func (pc *prometheusContorller) Release() {

}

func (pc *prometheusContorller) initRouter() {
	pc.router.AddRoute("/metrics", router.MethodGet, pc.metrics)
}
