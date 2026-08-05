package request

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(bytes), "\r\n")

	requestLine, err := parseRequestLine(lines[0])
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

func parseRequestLine(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Expected 3 request line parts but got %v", len(parts))
	}

	method := parts[0]
	if !slices.Contains([]string{"GET", "POST"}, method) {
		return nil, fmt.Errorf("Invalid method '%v'", method)
	}

	http, version, found := strings.Cut(parts[2], "/")
	if !found {
		return nil, fmt.Errorf("Couldn't find HTTP version in '%v'", parts[2])
	} else if http != "HTTP" {
		return nil, fmt.Errorf("Expected 'HTTP' but got '%v'", http)
	} else if version != "1.1" {
		return nil, fmt.Errorf("Only supports HTTP version '1.1' not '%v'", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: parts[1],
		HttpVersion:   version,
	}, nil
}
