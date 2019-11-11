package main

import (
	"os"
	"syscall"

	"github.com/YLonely/sysdig-monitor/server"

	"github.com/YLonely/sysdig-monitor/log"
)

var handledSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
}

func handleSignals(server server.Server, signals chan os.Signal, serverErrorC chan error) chan bool {
	done := make(chan bool, 1)
	go func() {
		select {
		case s := <-signals:
			log.L.WithField("signal", s).Debug("get signal")
		case err := <-serverErrorC:
			log.L.WithError(err).Error("server error")
		}
		// ignore the shutdown error
		server.Shutdown()
		close(done)
	}()
	return done
}
