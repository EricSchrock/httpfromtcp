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
)

type Writer struct {
	conn io.Writer
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{
		conn: conn,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
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
	for name, value := range headers {
		_, err := w.conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", name, value)))
		if err != nil {
			fmt.Errorf("Failed to write '%s: %s' header: %v", name, value, err)
		}
	}

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.conn.Write(append([]byte("\r\n"), p...))
}
