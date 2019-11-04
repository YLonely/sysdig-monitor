package server

import "context"

// Config containes params to start a server, only port now
type Config struct {
	port string
}

// Server is the interface of a monitor server
type Server interface {
	Start(ctx context.Context) chan error
}

type server struct {
	conf Config
}

func NewServer(conf Config) Server {
	return &server{conf: conf}
}

func (s *server) Start(ctx context.Context) chan error {
	
}
