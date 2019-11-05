package sysdig

import (
	"context"
	"encoding/json"
	"io"
	"os/exec"

	"github.com/YLonely/sysdig-monitor/log"
)

const binaryName = "sysdig"
const bufferSize = 10240
const formatString = "*%evt.num %evt.outputtime %evt.cpu %thread.tid %thread.vtid %proc.name %evt.dir %evt.type %evt.info " +
	//syscall
	"%syscall.type " +
	//container parts
	"%container.name %container.id " +
	//file or network parts
	"%fd.name %fd.type %evt.is_io_write %evt.is_io_read %evt.buffer %evt.buflen " +
	//performance
	"%evt.latency"

// Server starts sysdig and dispatch events
type Server interface {
	Subscribe() chan Event
	Start(ctx context.Context) (error, chan error)
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

func (ls *localServer) Start(ctx context.Context) (error, chan error) {
	if err := ls.preRrequestCheck(ctx); err != nil {
		log.L.WithError(err).Error("sysdig server pre check failed.")
		return err, nil
	}
	cmd := exec.CommandContext(ctx, binaryName, "-p", formatString, "-j", "container.name!=host")
	rd, err := cmd.StdoutPipe()
	if err != nil {
		return err, nil
	}
	ec, err := Monitor.Start(cmd)
	if err != nil {
		rd.Close()
		return err, nil
	}

	var (
		dec   = json.NewDecoder(rd)
		errCh = make(chan error, 1)
	)

	go func() {
		defer func() {
			close(errCh)
			rd.Close()
			Monitor.Wait(cmd, ec)
		}()
		for {
			var e Event
			if err := dec.Decode(&e); err != nil {
				if err != io.EOF {
					errCh <- err
				}
				return
			}
			for _, subscriber := range ls.subscribers {
				if !subscriber.closed {
					subscriber.c <- e
				}
			}
		}
	}()

	return nil, errCh
}

func (ls *localServer) preRrequestCheck(ctx context.Context) error {
	//try run sysdig
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, binaryName)
	var (
		ec  chan Exit
		err error
	)

	if ec, err = Monitor.Start(cmd); err != nil {
		cancel()
		return err
	}
	cancel()
	Monitor.Wait(cmd, ec)

	return nil
}
