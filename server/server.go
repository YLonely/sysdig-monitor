package server

import (
	"context"
	"net/http"
	"time"

	"github.com/YLonely/sysdig-monitor/server/controller"

	"github.com/gin-gonic/gin"

	"github.com/YLonely/sysdig-monitor/server/controller/container"
	"github.com/YLonely/sysdig-monitor/sysdig"
)

// Config containes params to start a server, only port now
type Config struct {
	Port string
}

// Server is the interface of a monitor server
type Server interface {
	Start(ctx context.Context) chan error
	Shutdown(ctx context.Context) error
}

type server struct {
	conf       Config
	httpServer *http.Server
}

func NewServer(conf Config) Server {
	return &server{conf: conf}
}

func (s *server) Start(ctx context.Context) chan error {
	errch := make(chan error, 1)
	sysdigServer := sysdig.NewServer()
	containerContorller, err := container.NewController(ctx, sysdigServer.Subscribe())
	if err != nil {
		errch <- err
		return errch
	}
	err, sysdigErrorCh := sysdigServer.Start(ctx)
	if err != nil {
		errch <- err
		return errch
	}
	gin.SetMode(gin.ReleaseMode)
	ginServer := gin.Default()
	initRoutes(ginServer, containerContorller) // may be more controller?
	s.httpServer = &http.Server{Addr: s.conf.Port, Handler: ginServer}
	httpServerErrorCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			httpServerErrorCh <- err
		}
	}()
	go func() {
		var e error
		select {
		case e = <-sysdigErrorCh:
			errch <- e
		case e = <-httpServerErrorCh:
			errch <- e
		}
	}()
	return errch
}

func (s *server) Shutdown(ctx context.Context) error {
	cctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	err := s.httpServer.Shutdown(cctx)
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
