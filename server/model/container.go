package model

import "time"

type Container struct {
	ID string

	SystemCalls

	FileSystem

	Network
}

type SystemCalls struct {
	// map system call name to SystemCall
	IndividualCalls map[string]SystemCall
	TotalCalls      int64
}

type FileSystem struct {
	// map file name to file
	AccessedFiles map[string]File
	// io calls whose latency is bigger than 1ms
	IOCalls1 []IOCall
	// io calls whose latency is bigger than 10ms
	IOCalls10 []IOCall
	// io calls whose latency is bigger than 100ms
	IOCalls100    []IOCall
	TotalReadIn   int64
	TotalWriteOut int64
}

type Network struct {
	ActiveConnections          map[ConnectionMeta]Connection
	TotalReadIn, TotalWriteOut int64
}

type SystemCall struct {
	Name string
	// total number of times it is invoked
	Calls     int64
	TotalTime time.Duration
}

type File struct {
	Name     string
	WriteOut int64
	ReadIn   int64
}

type Connection struct {
	Type string
	ConnectionMeta
	WriteOut int64
	ReadIn   int64
}

type ConnectionMeta struct {
	SourceIP   string
	DestIP     string
	SourcePort int
	DestPort   int
}

type IOCall struct {
	FileName string
	Latency  time.Duration
}
