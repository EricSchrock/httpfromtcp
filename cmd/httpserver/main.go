package main

import (
	"fmt"
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

func handler(w *response.Writer, req *request.Request) {
	var status response.StatusCode = response.StatusOK
	title := "200 OK"
	heading := "Success!"
	message := "Your request was an absolute banger."
	if target := req.RequestLine.RequestTarget; target == "/yourproblem" {
		status = response.StatusBadRequest
		title = "400 Bad Request"
		heading = "Bad Request"
		message = "Your request honestly kinda sucked."
	} else if target == "/myproblem" {
		status = response.StatusInternalServerError
		title = "500 Internal Server Error"
		heading = "Internal Server Error"
		message = "Okay, you know what? This one is on me."
	}

	body := fmt.Sprintf(`
	<html>
		<head>
			<title>%s</title>
		</head>
		<body>
			<h1>%s</h1>
			<p>%s</p>
		</body>
	</html>
	`, title, heading, message)

	err := w.WriteStatusLine(status)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(len(body))
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
