package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidSingleHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:12345\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:12345", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestValidSingleHeaderWithExtraWhitespace(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host:\t \tlocalhost:12345\t \t \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:12345", headers["Host"])
	assert.Equal(t, 29, n)
	assert.False(t, done)
}

func TestValidDone(t *testing.T) {
	headers := NewHeaders()
	data := []byte("\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)
}

func TestValidSecondHeader(t *testing.T) {
	headers := NewHeaders()
	headers["Host"] = "localhost:12345"
	data := []byte("User-Agent: curl/7.81.0\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:12345", headers["Host"])
	assert.Equal(t, "curl/7.81.0", headers["User-Agent"])
	assert.Equal(t, 25, n)
	assert.False(t, done)
}

func TestDuplicateSecondHeader(t *testing.T) {
	headers := NewHeaders()
	headers["Host"] = "localhost:12345"
	data := []byte("Host: localhost:12346\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:12345, localhost:12346", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestInvalidHeaderLeadingspace(t *testing.T) {
	headers := NewHeaders()
	data := []byte(" Host: localhost:12345\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestInvalidHeaderSpaceBeforeColon(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host : localhost:12345\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestInvalidWhitespaceInHeaderFieldName(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Ho st: localhost:12345\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
