package sysdig

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
)

const binaryName = "sysdig"
const bufferSize = 1024

// Server starts sysdig and dispatch events
type Server interface {
	Subscribe() chan Event
	Start(ctx context.Context) error
}

var _ Server = &localServer{}

type localServer struct {
	subscribers []*subscriber
}

// NewServer creates a server
func NewServer() Server {
	return &localServer{}
}

func (ls *localServer) Subscribe() chan Event {
	c := make(chan Event, bufferSize)
	ls.subscribers = append(ls.subscribers, &subscriber{c: c})
	return c
}

func (ls *localServer) Start(ctx context.Context) error {
	if err := ls.preRrequestCheck(); err != nil {
		return err
	}
	args := []string{"-pc", "-j", "container.id!=host"}
	cmd := exec.CommandContext(ctx, binaryName, args...)
	rd, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	ec, err := Monitor.Start(cmd)
	if err != nil {
		rd.Close()
		return err
	}

	var (
		dec = json.NewDecoder(rd)
	)

	go func() {
		defer func() {
			rd.Close()
			Monitor.Wait(cmd, ec)
		}()
		for {
			var e Event
			if err := dec.Decode(&e); err != nil {
				if err == io.EOF {
					return
				}
				e = Event{Type: "error"}
			}
			for _, subscriber := range ls.subscribers {
				if !subscriber.closed {
					subscriber.c <- e
				}
			}
		}
	}()

	return nil
}

func (ls *localServer) preRrequestCheck() error {
	//try run sysdig
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, binaryName)
	var (
		ec  chan Exit
		err error
	)

	if ec, err = Monitor.Start(cmd); err != nil {
		cancel()
		return errors.New("can not start sysdig")
	}

	cancel()

	Monitor.Wait(cmd, ec)

	return nil
}
