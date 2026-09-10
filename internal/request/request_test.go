package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestParseGetRequestLine(t *testing.T) {
	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestParseGetRequestLineThreeBytesAtAtime(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestParseGetRequestLineOneByteAtATime(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestParseGetRequestLineWithPath(t *testing.T) {
	r, err := RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestParsePostRequestLineWithPath(t *testing.T) {
	r, err := RequestFromReader(strings.NewReader("POST /coffee HTTP/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestParseRequestLineWithTooFewParts(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestParseRequestLineWithTooManyParts(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GET /cof fee HTTP/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestParseRequestLineWithInvalidMethod(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GOT /coffee HTTP/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestParseRequestLineWithUncapitalizedMethod(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("Get /coffee HTTP/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestParseRequestLineOutOfOrder(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("HTTP/1.1 Get /coffee\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestParseRequestLineWithTooFewVersionParts(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GET /coffee HTTP1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestParseRequestLineWithTooManyVersionParts(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestParseRequestLineWithInvalidFirstVersionPart(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GET /coffee HTTPS/1.1\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestParseRequestLineWithInvalidSecondVersionPart(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GET /coffee HTTP/1.0\r\nHost: localhost:12345\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestParseValidHeaders(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nhost: localhost:12345\r\nuser-agent: curl/7.81.0\r\naccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:12345", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])
}

func TestParseEmptyHeaders(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Empty(t, r.Headers)
}

func TestParseDuplicateHeaders(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nhost: localhost:12345\r\nhost: localhost:12346\r\n\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:12345, localhost:12346", r.Headers["host"])
}

func TestParseCaseInsensitiveHeaders(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:12345\r\n\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:12345", r.Headers["host"])
	assert.NotContains(t, r.Headers, "Host")
}

func TestParseMalformedHeader(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nhost localhost:12345\r\n\r\n"))
	require.Error(t, err)
}

func TestParseMissingEndOfHeaders(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nhost: localhost:12345\r\n"))
	require.Error(t, err)
}

func TestParseValidBody(t *testing.T) {
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"host: localhost:12345\r\n" +
			"content-length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))
}

func TestParseValidBodyNoContentLength(t *testing.T) {
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"host: localhost:12345\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body)) // parser drops body if no content-length header is present
}

func TestParseEmptyBody(t *testing.T) {
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"host: localhost:12345\r\n" +
			"content-length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))
}

func TestParseEmptyBodyNoContentLength(t *testing.T) {
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"host: localhost:12345\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))
}

func TestParseBodyShorterThanContentLength(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("POST /submit HTTP/1.1\r\nhost: localhost:12345\r\ncontent-length: 13\r\n\r\nhello world\n"))
	require.Error(t, err)
}

func TestParseBodyLongerThanContentLength(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("POST /submit HTTP/1.1\r\nhost: localhost:12345\r\ncontent-length: 13\r\n\r\nhello world!!\n"))
	require.Error(t, err)
}
