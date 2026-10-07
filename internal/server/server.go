package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/EricSchrock/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := Server{
		listener: listener,
	}

	go server.listen()

	return &server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if s.isClosed.Load() {
			continue
		} else if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	message := ""
	headers := response.GetDefaultHeaders(len(message))

	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
	}

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
	}

	_, err = conn.Write([]byte("\r\n" + message))
	if err != nil {
		log.Printf("Error writing message: %v", err)
	}

	conn.Close()
}
