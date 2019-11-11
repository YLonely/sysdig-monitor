package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/YLonely/sysdig-monitor/log"
	"github.com/YLonely/sysdig-monitor/server"

	"github.com/urfave/cli"
)

func main() {
	var port string

	app := cli.NewApp()
	app.Name = "sysdig-monitor"
	app.Usage = "Monitor using sysdig to trace all containers running on host."
	app.Version = "v0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "port",
			Value:       "8080",
			Usage:       "Specify a port to listen",
			Destination: &port,
		},
	}

	app.Action = func(c *cli.Context) error {
		conf := server.Config{Port: ":" + port}
		signals := make(chan os.Signal, 2048)
		ctx := context.Background()
		serv := server.NewServer(ctx, conf)
		errorC := serv.Start()
		done := handleSignals(serv, signals, errorC)
		signal.Notify(signals, handledSignals...)
		log.L.Info("sysdig-monitor successfully booted")
		<-done
		log.L.Info("sysdig-monitor exits")
		return nil
	}
	app.Run(os.Args)
}
