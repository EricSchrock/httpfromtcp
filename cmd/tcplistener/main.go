package main

import (
	"fmt"
	"log"
	"net"

	"github.com/EricSchrock/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error listening for TCP traffic: %s", err.Error())
	}
	defer listener.Close()
	fmt.Println("Listening for TCP traffic on", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting TCP connection: %s", err.Error())
			continue
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("Error reading request: %s", err.Error())
		}

		fmt.Println("Request line:")
		fmt.Println("- Method:", req.RequestLine.Method)
		fmt.Println("- Target:", req.RequestLine.RequestTarget)
		fmt.Println("- Version:", req.RequestLine.HttpVersion)
		fmt.Println()
		fmt.Println("Connection to", conn.RemoteAddr(), "closed")
	}
}
