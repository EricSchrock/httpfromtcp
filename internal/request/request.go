package request

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

type requestState int

const (
	initialized requestState = iota
	done
)

type Request struct {
	RequestLine RequestLine

	state requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		state: initialized,
	}

	readBuffer := make([]byte, 8)
	var parseBuffer []byte
	for req.state != done {
		n, err := reader.Read(readBuffer)
		if n > 0 {
			parseBuffer = append(parseBuffer, readBuffer[:n]...)
			bytesParsed, err := req.parse(parseBuffer)
			if err != nil {
				return nil, err
			}
			parseBuffer = parseBuffer[bytesParsed:]
		} else if err == io.EOF {
			req.state = done
			break
		} else if err != nil {
			return nil, err
		}
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == done {
		return 0, fmt.Errorf("Tried to parse a completed request")
	} else if r.state != initialized {
		return 0, fmt.Errorf("Unknown request state: %v", r.state)
	}

	requestLine, bytesParsed, err := parseRequestLine(string(data))
	if err != nil {
		return 0, err
	} else if bytesParsed > 0 {
		bytesParsed += 2 // account for \r\n
		r.state = done
		r.RequestLine = *requestLine
	}

	return bytesParsed, nil
}

func parseRequestLine(str string) (*RequestLine, int, error) {
	requestLine, _, found := strings.Cut(str, "\r\n")
	if !found {
		return nil, 0, nil
	}

	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, 0, fmt.Errorf("Expected 3 request line parts but got %v", len(parts))
	}

	method := parts[0]
	if !slices.Contains([]string{"GET", "POST"}, method) {
		return nil, 0, fmt.Errorf("Invalid method '%v'", method)
	}

	http, version, found := strings.Cut(parts[2], "/")
	if !found {
		return nil, 0, fmt.Errorf("Couldn't find HTTP version in '%v'", parts[2])
	} else if http != "HTTP" {
		return nil, 0, fmt.Errorf("Expected 'HTTP' but got '%v'", http)
	} else if version != "1.1" {
		return nil, 0, fmt.Errorf("Only supports HTTP version '1.1' not '%v'", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: parts[1],
		HttpVersion:   version,
	}, len(requestLine), nil
}
