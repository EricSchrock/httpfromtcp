package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
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

		ch := getLinesChannel(conn)

		for line := range ch {
			fmt.Println(line)
		}

		fmt.Println("Connection to", conn.RemoteAddr(), "closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer f.Close()

		buff := make([]byte, 8)
		line := ""
		for {
			n, err := f.Read(buff)
			if err != nil && err != io.EOF {
				log.Fatal(err)
			} else if err != io.EOF && n <= 0 {
				log.Fatal("No bytes read")
			} else if n <= 0 {
				break
			}

			parts := bytes.Split(buff[:n], []byte{'\n'})
			line += string(parts[0])
			for i := 1; i < len(parts); i++ {
				ch <- line
				line = string(parts[i])
			}

			if err == io.EOF {
				break
			}
		}
		if line != "" {
			ch <- line
		}
	}()

	return ch
}
