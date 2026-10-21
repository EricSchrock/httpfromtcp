package response

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func TestWriteChunkedBodyPreservesBody(t *testing.T) {
	body := []byte("abcdef")
	wantHash := sha256.Sum256(body)
	var output bytes.Buffer
	w := NewWriter(&output)
	w.state = headersWritten

	for i := 0; i < len(body); i += 3 {
		n, err := w.WriteChunkedBody(body[i : i+3])
		if err != nil {
			t.Fatal(err)
		}
		if n != 3 {
			t.Fatalf("wrote %d body bytes, want 3", n)
		}
	}
	if got := sha256.Sum256(body); got != wantHash {
		t.Fatalf("body changed during chunk writing: %q", body)
	}
	if got, want := output.String(), "3\r\nabc\r\n3\r\ndef\r\n"; got != want {
		t.Fatalf("chunked body = %q, want %q", got, want)
	}
}
