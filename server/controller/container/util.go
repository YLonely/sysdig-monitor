package container

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/YLonely/sysdig-monitor/log"
	"github.com/YLonely/sysdig-monitor/server/model"
)

const latency100 = time.Millisecond * 100
const latency10 = time.Millisecond * 10
const latency1 = time.Millisecond * 1

type containerEvent struct {
	eventType           string
	eventDir            string
	containerID         string
	containerName       string
	fdType              string
	fdName              string
	isIOWrite, isIORead bool
	bufferLen           int
	latency             time.Duration
	rawRes              int
	syscallType         string
	virtualtid          int
}

func processLoop(ctx context.Context, c *mutexContainer, ch chan containerEvent) error {
	var e containerEvent
	var err error
	for {
		select {
		case e = <-ch:
		case <-ctx.Done():
			return nil
		}
		if e.containerID != c.ID {
			return errors.New("event id mismatch")
		}
		// container exits
		if e.eventType == "procexit" && e.virtualtid == 1 {
			log.L.WithField("container-id", c.ID).Debug("container exits")
			return nil
		}
		if e.eventDir == "<" {
			c.m.Lock()
			if e.containerName != incompleteContainerName && c.Name == unknownContainerName {
				c.Name = e.containerName
			}
			if err = handleSysCall(c.Container, e); err != nil {
				log.L.WithError(err).WithField("container-id", c.ID).Error("syscall handler error")
			}
			if e.rawRes >= 0 && (e.isIORead || e.isIOWrite) {
				if err = handleIO(c.Container, e); err != nil {
					//if something wrong happens, just log it out
					log.L.WithField("container-id", c.ID).WithError(err).Error("io handle error")
				}
			} else {
				// may have some other handler?
				if err = handleNetwork(c.Container, e); err != nil {
					//log.L.WithField("container-id", c.ID).WithError(err).Error("network handler error")
				}
			}
			c.m.Unlock()
		}
	}
}

func handleSysCall(c *model.Container, e containerEvent) error {
	syscall := e.syscallType
	latency := e.latency
	if len(syscall) <= 0 {
		return nil
	}
	if _, exists := c.IndividualCalls[syscall]; !exists {
		c.IndividualCalls[syscall] = &model.SystemCall{Name: syscall}
	}
	call := c.IndividualCalls[syscall]
	call.Calls++
	call.TotalTime += latency
	c.SystemCalls.TotalCalls++
	return nil
}

func handleIO(c *model.Container, e containerEvent) error {
	if e.fdType == "file" {
		return handleFileIO(c, e)
	} else if e.fdType == "ipv4" || e.fdType == "ipv6" {
		return handleNetIO(c, e)
	}
	return nil
}

func handleNetIO(c *model.Container, e containerEvent) error {
	bufLen := e.bufferLen
	meta, err := connectionMeta(e.fdName)
	if err != nil {
		return err
	}
	// if event shows that a net io begins before "connect" or "accpet",
	// we just ignore the error sequence and add a new connection
	if _, exists := c.ActiveConnections[meta]; !exists {
		c.ActiveConnections[meta] = &model.Connection{Type: e.fdType}
	}
	conn := c.ActiveConnections[meta]
	if e.isIORead {
		conn.ReadIn += int64(bufLen)
		c.Network.TotalReadIn += int64(bufLen)
	} else if e.isIOWrite {
		conn.WriteOut += int64(bufLen)
		c.Network.TotalWriteOut += int64(bufLen)
	}
	return nil
}

func handleFileIO(c *model.Container, e containerEvent) error {
	fileName := e.fdName
	bufLen := e.bufferLen
	if _, exists := c.AccessedFiles[fileName]; !exists {
		c.AccessedFiles[fileName] = &model.File{Name: fileName}
	}
	file := c.AccessedFiles[fileName]
	if e.isIOWrite {
		file.WriteOut += int64(bufLen)
		c.FileSystem.TotalWriteOut += int64(bufLen)
	} else if e.isIORead {
		file.ReadIn += int64(bufLen)
		c.FileSystem.TotalReadIn += int64(bufLen)
	}
	latency := e.latency
	latency /= time.Millisecond
	if !strings.HasPrefix(e.fdName, "/dev/") {
		iocall := &model.IOCall{FileName: fileName, Latency: e.latency}
		if latency > latency100 {
			c.IOCalls100 = append(c.IOCalls100, iocall)
		}
		if latency > latency10 {
			c.IOCalls10 = append(c.IOCalls10, iocall)
		}
		if latency > latency1 {
			c.IOCalls1 = append(c.IOCalls1, iocall)
		}
	}
	return nil
}

func handleNetwork(c *model.Container, e containerEvent) error {
	var (
		meta model.ConnectionMeta
		err  error
	)
	if e.eventType == "connect" || e.eventType == "accept" {
		if meta, err = connectionMeta(e.fdName); err != nil {
			return err
		}
		if _, exists := c.ActiveConnections[meta]; !exists {
			c.ActiveConnections[meta] = &model.Connection{Type: e.fdType}
		}
	} else if e.eventType == "close" && !strings.HasPrefix(e.fdName, "/") {
		if meta, err = connectionMeta(e.fdName); err != nil {
			return err
		}
		// should return nonexists error?
		if _, exists := c.ActiveConnections[meta]; exists {
			delete(c.ActiveConnections, meta)
		}
	}
	return nil
}

func connectionMeta(fdname string) (model.ConnectionMeta, error) {
	parts := strings.Split(fdname, "->")
	meta := model.ConnectionMeta{}
	if len(parts) != 2 {
		return meta, fmt.Errorf("wrong connection meta format:%v", fdname)
	}
	source, dest := parts[0], parts[1]

	var err error
	if meta.SourceIP, meta.SourcePort, err = splitAddress(source); err != nil {
		return meta, err
	}
	if meta.DestIP, meta.DestPort, err = splitAddress(dest); err != nil {
		return meta, err
	}
	return meta, nil
}

func splitAddress(address string) (string, int, error) {
	var (
		portStart, port int
		err             error
	)
	if len(address) <= 1 {
		return "", -1, errors.New("empty address")
	}
	for portStart = len(address) - 1; portStart >= 0 && address[portStart] != ':'; portStart-- {
	}
	if portStart <= 0 {
		return "", -1, errors.New("no port address")
	}
	if port, err = strconv.Atoi(address[portStart+1:]); err != nil {
		return "", -1, fmt.Errorf("wrong address format:%v", address)
	}
	return address[:portStart], port, nil
}
