package request

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/EricSchrock/httpfromtcp/internal/headers"
)

type requestState int

const (
	requestInitialized requestState = iota
	parsingHeaders
	parsingBody
	parsingComplete
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte

	state requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		state:   requestInitialized,
		Headers: headers.NewHeaders(),
	}

	readBuffer := make([]byte, 8)
	var parseBuffer []byte
	for req.state != parsingComplete {
		n, err := reader.Read(readBuffer)
		if (err != nil) && (err != io.EOF) {
			return nil, err
		} else if n > 0 {
			parseBuffer = append(parseBuffer, readBuffer[:n]...)
			bytesParsed, err := req.parse(parseBuffer)
			if err != nil {
				return nil, err
			}
			parseBuffer = parseBuffer[bytesParsed:]
		}

		if (err == io.EOF) && (req.state != parsingComplete) {
			return nil, fmt.Errorf("Incomplete request received")
		}
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == parsingComplete {
		return 0, fmt.Errorf("Tried to parse a completed request")
	}

	totalBytesParsed := 0
	if r.state == requestInitialized {
		requestLine, bytesParsed, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		} else if bytesParsed <= 0 {
			return totalBytesParsed, nil
		}

		r.state = parsingHeaders
		r.RequestLine = *requestLine
		totalBytesParsed += bytesParsed
	}

	for r.state == parsingHeaders {
		bytesParsed, final, err := r.Headers.Parse(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		} else if bytesParsed <= 0 {
			return totalBytesParsed, nil
		} else if final {
			r.state = parsingBody
		}
		totalBytesParsed += bytesParsed
	}

	if r.state == parsingBody {
		value, found := r.Headers.Get("content-length")
		if !found {
			r.state = parsingComplete // assume there is no body
			return totalBytesParsed, nil
		}

		content_length, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
		}

		r.Body = append(r.Body, data[totalBytesParsed:]...)
		totalBytesParsed = len(data)

		if len(r.Body) > content_length {
			return 0, fmt.Errorf("Body length is >= '%v' bytes but content-length='%v'", len(r.Body), content_length)
		} else if len(r.Body) == content_length {
			r.state = parsingComplete
		}
	}

	return totalBytesParsed, nil
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
	}, len(requestLine) + 2, nil // acount for \r\n
}
