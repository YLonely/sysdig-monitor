package model

import (
	"fmt"
	"errors"
	"strings"
	"time"
)

type Digest string

type Container struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	SystemCalls

	FileSystem

	Network
}

func NewContainer(id, name string, layerData map[string]string) (*Container, error) {
	lowerLayers, exists := layerData["LowerDir"]
	if !exists {
		return nil, errors.New("cant find lowerdir in layerdata")
	}
	upperLayer, exists := layerData["UpperDir"]
	if !exists {
		return nil, errors.New("cant find upperdir in layerdata")
	}
	layers := strings.Split(lowerLayers, ":")
	if len(layers) == 0 {
		return nil, errors.New("no lowerlayers exists")
	}
	layers = append([]string{upperLayer}, layers...)
	system := SystemCalls{IndividualCalls: map[string]*SystemCall{}}
	fileSys := FileSystem{AccessedLayers: map[string]*LayerInfo{}, LayersInOrder: layers, AccessedFiles: map[string]*File{}}
	for _, l := range layers {
		fileSys.AccessedLayers[l] = NewLayerInfo(l)
	}

	net := Network{ActiveConnections: map[ConnectionMeta]*Connection{}}
	return &Container{ID: id, Name: name, SystemCalls: system, FileSystem: fileSys, Network: net}, nil
}

func convertToDigest(layers []string) ([]Digest, error) {
	digests := make([]Digest, 0, len(layers))
	for _, layer := range layers {
		dirs := strings.Split(layer, "/")
		if len(dirs) < 2 {
			return nil, fmt.Errorf("depth of the layer %v is less than 2", layer)
		}
		digests = append(digests, Digest(dirs[len(dirs)-2]))
	}
	return digests, nil
}

func NewLayerInfo(dir string) *LayerInfo {
	return &LayerInfo{AccessedFiles: map[string]*File{}, Dir: dir}
}

type SystemCalls struct {
	// map system call name to SystemCall
	IndividualCalls map[string]*SystemCall `json:"individual_calls"`
	TotalCalls      int64                  `json:"total_calls"`
}

type FileSystem struct {
	// map file name to file
	AccessedFiles map[string]*File `json:"-"`
	// map layer path to layer info
	AccessedLayers map[string]*LayerInfo `json:"-"`
	// we have to record this to ensure the file search order
	LayersInOrder []string `json:"-"`
	// io calls whose latency is bigger than 1ms
	IOCalls1 []*IOCall `json:"io_calls_more_than_1ms"`
	// io calls whose latency is bigger than 10ms
	IOCalls10 []*IOCall `json:"io_calls_more_than_10ms"`
	// io calls whose latency is bigger than 100ms
	IOCalls100    []*IOCall `json:"io_calls_more_than_100ms"`
	TotalReadIn   int64     `json:"file_total_read_in"`
	TotalWriteOut int64     `json:"file_total_write_out"`
}

type Network struct {
	ActiveConnections map[ConnectionMeta]*Connection `json:"-"`
	TotalReadIn       int64                          `json:"net_total_read_in"`
	TotalWriteOut     int64                          `json:"net_total_wirte_out"`
}

type SystemCall struct {
	Name string `json:"-"`
	// total number of times it is invoked
	Calls     int64         `json:"calls"`
	TotalTime time.Duration `json:"total_time"`
}

type LayerInfo struct {
	Dir           string           `json:"dir"`
	AccessedFiles map[string]*File `json:"accessed_files"`
	WriteOut      int64            `json:"layer_write_out"`
	ReadIn        int64            `json:"layer_read_in"`
}

type File struct {
	Name     string     `json:"-"`
	WriteOut int64      `json:"write_out"`
	ReadIn   int64      `json:"read_in"`
	Layer    *LayerInfo `json:"-"`
}

type Connection struct {
	// ipv4 or ipv6
	Type     string `json:"type"`
	WriteOut int64  `json:"write_out"`
	ReadIn   int64  `json:"read_in"`
}

type ConnectionMeta struct {
	SourceIP   string `json:"source_ip"`
	DestIP     string `json:"dest_ip"`
	SourcePort int    `json:"source_port"`
	DestPort   int    `json:"dest_port"`
}

type IOCall struct {
	FileName string        `json:"file_name"`
	Latency  time.Duration `json:"latency"`
}
