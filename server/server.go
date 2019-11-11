package server

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/YLonely/sysdig-monitor/server/controller/prometheus"

	"github.com/YLonely/sysdig-monitor/server/controller"

	"github.com/gin-gonic/gin"

	"github.com/YLonely/sysdig-monitor/log"
	"github.com/YLonely/sysdig-monitor/server/controller/container"
)

// Config containes params to start a server, only port now
type Config struct {
	Port string
}

// Server is the interface of a monitor server
type Server interface {
	Start() chan error
	Shutdown() error
}

type server struct {
	conf        Config
	httpServer  *http.Server
	cancle      context.CancelFunc
	ctx         context.Context
	controllers []controller.Controller
}

func NewServer(ctx context.Context, conf Config) Server {
	res := &server{conf: conf}
	ctx, cancle := context.WithCancel(ctx)
	res.cancle = cancle
	res.ctx = ctx
	return res
}

func (s *server) Start() chan error {
	errch := make(chan error, 1)
	containerContorller, err := container.NewController(s.ctx, errch)
	promContorller := prometheus.NewController(s.ctx)
	if err != nil {
		errch <- err
		return errch
	}
	s.controllers = append(s.controllers, containerContorller, promContorller)
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	ginServer := gin.Default()
	initRoutes(ginServer, s.controllers...) // may be more controller?
	s.httpServer = &http.Server{Addr: s.conf.Port, Handler: ginServer}
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errch <- err
		}
	}()
	log.L.Info("web server start. Listening on port " + s.conf.Port)
	return errch
}

func (s *server) Shutdown() error {
	ctx, cancel := context.WithTimeout(s.ctx, time.Second*3)
	defer cancel()
	err := s.httpServer.Shutdown(ctx)
	s.cancle()
	for _, c := range s.controllers {
		c.Release()
	}
	return err
}

func initRoutes(ginServer *gin.Engine, controllers ...controller.Controller) {
	for _, controller := range controllers {
		routes := controller.BindedRoutes()
		for _, route := range routes {
			ginServer.Handle(route.Method(), route.Path(), gin.HandlerFunc(route.Handler()))
		}
	}
}
