package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	if target := req.RequestLine.RequestTarget; strings.HasPrefix(target, "/httpbin/") {
		proxy(w, req)
	} else if target == "/yourproblem" {
		defaultResponse(w, response.StatusBadRequest)
	} else if target == "/myproblem" {
		defaultResponse(w, response.StatusInternalServerError)
	} else {
		defaultResponse(w, response.StatusOK)
	}
}

func proxy(w *response.Writer, req *request.Request) {
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	target := "https://httpbingo.org/" + path

	resp, err := http.Get(target)
	if err != nil {
		log.Printf("Request to '%s' failed: %v", target, err)
		defaultResponse(w, response.StatusBadGateway)
	}
	defer resp.Body.Close()

	err = w.WriteStatusLine(response.StatusOK)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(0)
	headers.Remove("Content-Length")
	headers.Add("Transfer-Encoding", "chunked")
	headers.Add("Trailers", "X-Content-SHA256")
	headers.Add("Trailers", "X-Content-Length")

	err = w.WriteHeaders(headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	body := make([]byte, 0)
	chunk := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(chunk)

		if n > 0 {
			_, err = w.WriteChunkedBody(chunk)
			if err != nil {
				log.Printf("Error writing chunk: %v", err)
				return
			}

			body = append(body, chunk[:n]...)
		}

		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Error reading chunk: %v", err)
			return
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		log.Printf("Error writing final chunk: %v", err)
		return
	}

	headers.Add("X-Content-SHA256", fmt.Sprintf("%x", sha256.Sum256(body)))
	headers.Add("X-Content-Length", fmt.Sprintf("%d", len(body)))

	err = w.WriteTrailers(headers)
	if err != nil {
		log.Printf("Error writing trailers: %v", err)
		return
	}
}

func defaultResponse(w *response.Writer, status response.StatusCode) {
	var title string
	var heading string
	var message string
	if status == response.StatusOK {
		title = "200 OK"
		heading = "Success!"
		message = "Your request was an absolute banger."
	} else if status == response.StatusBadRequest {
		title = "400 Bad Request"
		heading = "Bad Request"
		message = "Your request honestly kinda sucked."
	} else if status == response.StatusInternalServerError {
		title = "500 Internal Server Error"
		heading = "Internal Server Error"
		message = "Okay, you know what? This one is on me."
	} else if status == response.StatusBadGateway {
		title = "502 Bad Gateway"
		heading = "Bad Gateway"
		message = ""
	} else {
		message = fmt.Sprintf("Unsupported response status: %v", status)
		log.Print(message)
		status = response.StatusInternalServerError
		title = "500 Internal Server Error"
		heading = "Internal Server Error"
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
