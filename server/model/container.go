package model

import "time"

type Container struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	SystemCalls

	FileSystem

	Network
}

func NewContainer(id, name string) *Container {
	system := SystemCalls{IndividualCalls: map[string]*SystemCall{}}
	fileSys := FileSystem{AccessedFiles: map[string]*File{}}
	net := Network{ActiveConnections: map[ConnectionMeta]*Connection{}}
	return &Container{ID: id, Name: name, SystemCalls: system, FileSystem: fileSys, Network: net}
}

type SystemCalls struct {
	// map system call name to SystemCall
	IndividualCalls map[string]*SystemCall `json:"individual_calls"`
	TotalCalls      int64                  `json:"total_calls"`
}

type FileSystem struct {
	// map file name to file
	AccessedFiles map[string]*File `json:"accessed_files"`
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

type File struct {
	Name     string `json:"-"`
	WriteOut int64  `json:"write_out"`
	ReadIn   int64  `json:"read_in"`
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
