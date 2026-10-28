package main

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/EricSchrock/httpfromtcp/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	s, _ := server.Serve(port, handler)
	defer s.Close()
	m.Run()
}

func TestGet(t *testing.T) {
	target := fmt.Sprintf("http://localhost:%d", port)
	resp, err := http.Get(target)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEqual(t, int64(-1), resp.ContentLength)
	assert.True(t, resp.Close)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
	assert.Contains(t, string(body), "absolute banger")
}

func TestGetBadRequest(t *testing.T) {
	target := fmt.Sprintf("http://localhost:%d/yourproblem", port)
	resp, err := http.Get(target)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.NotEqual(t, int64(-1), resp.ContentLength)
	assert.True(t, resp.Close)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
	assert.Contains(t, string(body), "kinda sucked")
}

func TestGetServerError(t *testing.T) {
	target := fmt.Sprintf("http://localhost:%d/myproblem", port)
	resp, err := http.Get(target)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.NotEqual(t, int64(-1), resp.ContentLength)
	assert.True(t, resp.Close)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
	assert.Contains(t, string(body), "This one is on me.")
}

func TestChunkedProxyGet(t *testing.T) {
	target := fmt.Sprintf("http://localhost:%d/httpbin/stream/100", port)
	resp, err := http.Get(target)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int64(-1), resp.ContentLength)
	assert.True(t, resp.Close)
	assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	assert.Contains(t, string(body), "httpbingo.org")
	assert.Equal(t, "chunked", resp.TransferEncoding[0])
}

func TestChunkedProxyGetTrailers(t *testing.T) {
	target := fmt.Sprintf("http://localhost:%d/httpbin/range/4096", port)
	resp, err := http.Get(target)
	require.NoError(t, err)
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int64(-1), resp.ContentLength)
	assert.True(t, resp.Close)
	assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	assert.Equal(t, "chunked", resp.TransferEncoding[0])
	assert.Equal(t, "4096", resp.Trailer.Get("X-Content-Length"))
	assert.Equal(t, "bc45051ac426475f459ec0b0c88a6646d037b8dfb1b9fa3ca3ef9203ce33e283", resp.Trailer.Get("X-Content-SHA256"))
}
