package response

import (
	"fmt"
	"io"

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
		return fmt.Errorf("Invalid response writer state: %v", w.state)
	}

	statusMap := map[StatusCode]string{
		StatusOK:                  "OK",
		StatusBadRequest:          "Bad Request",
		StatusInternalServerError: "Internal Server Error",
	}

	reasonPhrase, ok := statusMap[statusCode]
	if !ok {
		return fmt.Errorf("Invalid status code: %v", statusCode)
	}

	_, err := w.conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)))
	if err != nil {
		return fmt.Errorf("Failed to write status line: %v", err)
	}

	w.state = statusLineWritten
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers["Content-Length"] = fmt.Sprintf("%d", contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/html"

	return headers
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != statusLineWritten {
		return fmt.Errorf("Invalid response writer state: %v", w.state)
	}

	for name, value := range headers {
		_, err := w.conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", name, value)))
		if err != nil {
			return fmt.Errorf("Failed to write '%s: %s' header: %v", name, value, err)
		}
	}

	_, err := w.conn.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("Failed to write trailing new line after headers: %v", err)
	}

	w.state = headersWritten
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != headersWritten {
		return 0, fmt.Errorf("Invalid response writer state: %v", w.state)
	}

	w.state = bodyWritten
	return w.conn.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != headersWritten {
		return 0, fmt.Errorf("Invalid response writer state: %v", w.state)
	}

	n, err := w.conn.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
	if err != nil {
		return n, err
	}

	return w.conn.Write(append(p, []byte("\r\n")...))
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != headersWritten {
		return 0, fmt.Errorf("Invalid response writer state: %v", w.state)
	}

	w.state = bodyWritten
	return w.conn.Write([]byte("0\r\n\r\n"))
}
