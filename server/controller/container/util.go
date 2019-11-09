package container

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/YLonely/sysdig-monitor/errdefs"
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
	containerIDMetric.WithLabelValues(c.ID).Set(1)
	depth := 1
	for _, l := range c.LayersInOrder {
		containerLayerDir.WithLabelValues(c.ID, l).Set(float64(depth))
		depth++
	}
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
			cleanUpMetrics(c.Container)
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
				if err = handleNetwork(c.Container, e); err != nil && !errdefs.IsErrWrongFormat(err) {
					log.L.WithField("container-id", c.ID).WithError(err).Warning("network handler error")
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
	var (
		call   *model.SystemCall
		exists bool
	)
	if call, exists = c.IndividualCalls[syscall]; !exists {
		call = &model.SystemCall{Name: syscall}
		c.IndividualCalls[syscall] = call
	}
	call.Calls++
	call.TotalTime += latency
	c.SystemCalls.TotalCalls++
	systemCallCount.WithLabelValues(c.ID, syscall).Set(float64(call.Calls))
	systemCallTotalLatency.WithLabelValues(c.ID, syscall).Set(float64(call.TotalTime) / float64(time.Second))
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
	var (
		conn   *model.Connection
		exists bool
	)
	if conn, exists = c.ActiveConnections[meta]; !exists {
		conn = &model.Connection{Type: e.fdType}
		c.ActiveConnections[meta] = conn
	}
	sip, sport, dip, dport := meta.SourceIP, strconv.Itoa(meta.SourcePort), meta.DestIP, strconv.Itoa(meta.DestPort)
	if e.isIORead {
		conn.ReadIn += int64(bufLen)
		c.Network.TotalReadIn += int64(bufLen)
		containerActiveConnectionRead.WithLabelValues(c.ID, sip, sport, dip, dport).Set(float64(conn.ReadIn))
	} else if e.isIOWrite {
		conn.WriteOut += int64(bufLen)
		c.Network.TotalWriteOut += int64(bufLen)
		containerActiveConnectionWrite.WithLabelValues(c.ID, sip, sport, dip, dport).Set(float64(conn.WriteOut))
	}
	return nil
}

func handleFileIO(c *model.Container, e containerEvent) error {
	fileName := e.fdName
	bufLen := e.bufferLen
	var (
		file   *model.File
		exists bool
	)

	if file, exists = c.AccessedFiles[fileName]; !exists {
		file = &model.File{Name: fileName}
		err := attachToLayer(c, file)
		if err != nil {
			return nil
		}
		c.AccessedFiles[fileName] = file
	}
	layerDir := file.Layer.Dir
	if e.isIOWrite {
		file.WriteOut += int64(bufLen)
		file.Layer.WriteOut += int64(bufLen)
		c.FileSystem.TotalWriteOut += int64(bufLen)
		containerLayerFileWrite.WithLabelValues(c.ID, layerDir, fileName).Set(float64(file.WriteOut))
	} else if e.isIORead {
		file.Layer.ReadIn += int64(bufLen)
		file.ReadIn += int64(bufLen)
		c.FileSystem.TotalReadIn += int64(bufLen)
		containerLayerFileRead.WithLabelValues(c.ID, layerDir, fileName).Set(float64(file.ReadIn))
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
		sip, sport, dip, dport := meta.SourceIP, strconv.Itoa(meta.SourcePort), meta.DestIP, strconv.Itoa(meta.DestPort)
		if _, exists := c.ActiveConnections[meta]; exists {
			delete(c.ActiveConnections, meta)
			containerActiveConnectionRead.DeleteLabelValues(c.ID, sip, sport, dip, dport)
			containerActiveConnectionWrite.DeleteLabelValues(c.ID, sip, sport, dip, dport)
		}
	}
	return nil
}

func connectionMeta(fdname string) (model.ConnectionMeta, error) {
	parts := strings.Split(fdname, "->")
	meta := model.ConnectionMeta{}
	if len(parts) != 2 {
		return meta, errdefs.NewErrWrongFormat(fdname)
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
		return "", -1, errdefs.NewErrWrongFormat(address)
	}
	for portStart = len(address) - 1; portStart >= 0 && address[portStart] != ':'; portStart-- {
	}
	if portStart <= 0 {
		return "", -1, errdefs.NewErrWrongFormat(address)
	}
	if port, err = strconv.Atoi(address[portStart+1:]); err != nil {
		return "", -1, errdefs.NewErrWrongFormat(address)
	}
	return address[:portStart], port, nil
}

func attachToLayer(c *model.Container, file *model.File) error {
	fileName := file.Name
	for _, dir := range c.LayersInOrder {
		if _, err := os.Stat(dir + fileName); err == nil {
			c.AccessedLayers[dir].AccessedFiles[fileName] = file
			file.Layer = c.AccessedLayers[dir]
			return nil
		}
	}
	return fmt.Errorf("cant find file %v in any of lower layers", fileName)
}

func cleanUpMetrics(container *model.Container) {
	// clean up connection metrics
	for meta, _ := range container.ActiveConnections {
		sip, sport, dip, dport := meta.SourceIP, strconv.Itoa(meta.SourcePort), meta.DestIP, strconv.Itoa(meta.DestPort)
		containerActiveConnectionRead.DeleteLabelValues(container.ID, sip, sport, dip, dport)
		containerActiveConnectionWrite.DeleteLabelValues(container.ID, sip, sport, dip, dport)
	}

	// clean up layer file metrics and layer metric
	for layerDir, layerInfo := range container.AccessedLayers {
		for fileName, _ := range layerInfo.AccessedFiles {
			containerLayerFileRead.DeleteLabelValues(container.ID, layerDir, fileName)
			containerLayerFileWrite.DeleteLabelValues(container.ID, layerDir, fileName)
		}
		containerLayerDir.DeleteLabelValues(container.ID, layerDir)
	}

	// clean up system call count and latency metrics
	for syscall, _ := range container.IndividualCalls {
		systemCallCount.DeleteLabelValues(container.ID, syscall)
		systemCallTotalLatency.DeleteLabelValues(container.ID, syscall)
	}

	// delete this container
	containerIDMetric.DeleteLabelValues(container.ID)
}
