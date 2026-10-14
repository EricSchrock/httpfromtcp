package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EricSchrock/httpfromtcp/internal/request"
	"github.com/EricSchrock/httpfromtcp/internal/response"
	"github.com/EricSchrock/httpfromtcp/internal/server"
)

const port = 12345

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	if target := req.RequestLine.RequestTarget; target == "/yourproblem" {
		return &server.HandlerError{
			Status:  response.StatusBadRequest,
			Message: "Your problem is not my problem\n",
		}
	} else if target == "/myproblem" {
		return &server.HandlerError{
			Status:  response.StatusInternalServerError,
			Message: "Woopsie, my bad\n",
		}
	}

	w.Write([]byte("All good, frfr\n"))

	return nil
}
