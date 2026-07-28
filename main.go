package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal(err)
	}
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
			fmt.Printf("read: %s\n", line)
			line = string(parts[i])
		}

		if err == io.EOF {
			break
		}
	}
	fmt.Printf("read: %s\n", line)
}
