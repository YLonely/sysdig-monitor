package main

import (
	"github.com/urfave/cli"
)

func main() {
	var port string

	app := cli.NewApp()
	app.Name = "sysdig-monitor"
	app.Usage = "Monitor using sysdig to trace all containers running on host."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "port",
			Value:       "8080",
			Usage:       "Specify a port to listen",
			Destination: &port,
		},
	}

	
}
