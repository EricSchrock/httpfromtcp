package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/EricSchrock/httpfromtcp/internal/request"
	"github.com/EricSchrock/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	isClosed atomic.Bool
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := Server{
		listener: listener,
		handler:  handler,
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
		if err == net.ErrClosed || s.isClosed.Load() {
			return
		} else if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	w := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err == nil {
		s.handler(w, req)
		return
	}

	body := err.Error()

	err = w.WriteStatusLine(response.StatusBadRequest)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(len(body))
	headers["Content-Type"] = "text/plain"
	err = w.WriteHeaders(headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	_, err = w.WriteBody([]byte(body))
	if err != nil {
		log.Printf("Error writing body: %v", err)
		return
	}
}
