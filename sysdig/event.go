package sysdig

// Event represents a container event in sysdig
type Event struct {
	ID              string `json:"container.id"`
	Name            string `json:"container.name"`
	CPUId           int    `json:"evt.cpu"`
	Dir             string `json:"evt.dir"`
	Info            string `json:"evt.info"`
	Num             int    `json:"evt.num"`
	TimeStamp       int64  `json:"evt.outputtime"`
	Type            string `json:"evt.type"`
	ProcName        string `json:"proc.name"`
	ThreadID        int    `json:"thread.tid"`
	ThreadVirtualID int    `json:"thread.vid"`
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