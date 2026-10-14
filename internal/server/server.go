package server

import (
	"bytes"
	"fmt"
	"io"
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

type HandlerError struct {
	Status  response.StatusCode
	Message string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

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
	defer conn.Close()

	var status response.StatusCode
	var message string

	req, err := request.RequestFromReader(conn)
	if err != nil {
		status = response.StatusBadRequest
		message = err.Error()
	} else {
		var buf bytes.Buffer
		handlerError := s.handler(&buf, req)
		if handlerError != nil {
			status = handlerError.Status
			message = handlerError.Message
		} else {
			status = response.StatusOK
			message = buf.String()
		}
	}

	err = response.WriteStatusLine(conn, status)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(len(message))
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	_, err = conn.Write([]byte("\r\n" + message))
	if err != nil {
		log.Printf("Error writing message: %v", err)
		return
	}
}
