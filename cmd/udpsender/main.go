package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:12345")
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err.Error())
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("Failed to dial UDP address: %v", err.Error())
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Failed to read input: %v", err.Error())
		}

		_, err = conn.Write([]byte(text))
		if err != nil {
			log.Printf("Failed to write to connection: %v", err.Error())
		}
	}
}
