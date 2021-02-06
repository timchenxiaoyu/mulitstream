package mulitstream

import (
	"io"
)

type Server struct {
	s *Session
}

func NewServer(conn io.ReadWriteCloser, c *Config) *Server {
	server := &Server{s: NewSession(conn)}
	return server

}

func (s *Server) Stop() {
	s.s.Close()
}
