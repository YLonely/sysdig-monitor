package sysdig

import "time"

// Event represents a container event in sysdig
type Event struct {
	ContainerID     string        `json:"container.id"`
	ContainerName   string        `json:"container.name"`
	EventCPU        int           `json:"evt.cpu"`
	EventDir        string        `json:"evt.dir"`
	EventInfo       string        `json:"evt.info"`
	EventNum        int           `json:"evt.num"`
	EventOutputTime int64         `json:"evt.outputtime"`
	EventType       string        `json:"evt.type"`
	FdName          string        `json:"fd.name"`
	FdType          string        `json:"fd.type"`
	EventIsIORead   bool          `json:"evt.is_io_read"`
	EventIsIOWrite  bool          `json:"evt.is_io_write"`
	EventBuffer     string        `json:"evt.buffer"`
	EventBuflen     int           `json:"evt.buflen"`
	ProcName        string        `json:"proc.name"`
	ThreadID        int           `json:"thread.tid"`
	ThreadVirtualID int           `json:"thread.vid"`
	EventLatency    time.Duration `json:"evt.latency"`
	RawRes          int           `json:"evt.rawres"`
	SyscallType     string        `json:"syscall.type"`
}

type subscriber struct {
	c      chan Event
	closed bool
}

func (s *subscriber) close() {
	if s.closed {
		return
	}
	close(s.c)
	s.closed = true
	return
}
