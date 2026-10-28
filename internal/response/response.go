package response

import (
	"fmt"
	"io"
	"strings"

	"github.com/EricSchrock/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  = 200
	StatusBadRequest          = 400
	StatusInternalServerError = 500
	StatusBadGateway          = 502
)

type writerState int

const (
	writerInitialized writerState = iota
	statusLineWritten
	headersWritten
	bodyWritten
	trailersWritten
)

type Writer struct {
	state writerState
	conn  io.Writer
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{
		state: writerInitialized,
		conn:  conn,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerInitialized {
		return fmt.Errorf("invalid response writer state: %v", w.state)
	}

	statusMap := map[StatusCode]string{
		StatusOK:                  "OK",
		StatusBadRequest:          "Bad Request",
		StatusInternalServerError: "Internal Server Error",
	}

	reasonPhrase, ok := statusMap[statusCode]
	if !ok {
		return fmt.Errorf("invalid status code: %v", statusCode)
	}

	_, err := fmt.Fprintf(w.conn, "HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	if err != nil {
		return fmt.Errorf("failed to write status line: %v", err)
	}

	w.state = statusLineWritten
	return nil
}

func GetDefaultHeaders(contentLen int, contentType string) headers.Headers {
	headers := headers.NewHeaders()

	headers.Add("Content-Length", fmt.Sprintf("%d", contentLen))
	headers.Add("Connection", "close")
	headers.Add("Content-Type", contentType)

	return headers
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != statusLineWritten {
		return fmt.Errorf("invalid response writer state: %v", w.state)
	}

	for name, value := range headers {
		_, err := fmt.Fprintf(w.conn, "%s: %s\r\n", name, value)
		if err != nil {
			return fmt.Errorf("failed to write '%s: %s' header: %v", name, value, err)
		}
	}

	_, err := w.conn.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("failed to write trailing new line after headers: %v", err)
	}

	w.state = headersWritten
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != headersWritten {
		return 0, fmt.Errorf("invalid response writer state: %v", w.state)
	}

	w.state = bodyWritten
	return w.conn.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != headersWritten {
		return 0, fmt.Errorf("invalid response writer state: %v", w.state)
	}

	n, err := fmt.Fprintf(w.conn, "%x\r\n", len(p))
	if err != nil {
		return n, err
	}

	n, err = w.conn.Write(p)
	if err != nil {
		return n, err
	}
	_, err = w.conn.Write([]byte("\r\n"))
	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != headersWritten {
		return 0, fmt.Errorf("invalid response writer state: %v", w.state)
	}

	w.state = bodyWritten
	return w.conn.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(headers headers.Headers) error {
	if w.state != bodyWritten {
		return fmt.Errorf("invalid response writer state: %v", w.state)
	}

	trailers, found := headers.Get("Trailer")
	if !found {
		return fmt.Errorf("trailers header not found")
	}

	for _, name := range strings.Split(trailers, ", ") {
		value, found := headers.Get(name)
		if !found {
			return fmt.Errorf("expected a '%s' trailer but found none", name)
		}

		_, err := fmt.Fprintf(w.conn, "%s: %s\r\n", name, value)
		if err != nil {
			return fmt.Errorf("failed to write '%s: %s' trailer: %v", name, value, err)
		}
	}

	_, err := w.conn.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("failed to write trailing new line after trailers: %v", err)
	}

	w.state = trailersWritten
	return nil
}
