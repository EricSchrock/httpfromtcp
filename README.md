# HTTP From TCP

HTTP/1.1 parser and server implementation from the [boot.dev](https://www.boot.dev/courses/learn-http-protocol-golang) HTTP protocol course.

## Specifications

1. [RFC 9110: HTTP Semantics](https://datatracker.ietf.org/doc/html/rfc9110)
2. [RFC 9112: HTTP/1.1](https://datatracker.ietf.org/doc/html/rfc9112)

## Installation

### Test Framework

```sh
go get -u github.com/stretchr/testify/assert
```

### Server Assets

```
mkdir assets && curl -o assets/vim.mp4 https://storage.googleapis.com/qvault-webapp-dynamic-assets/lesson_videos/vim-vs-neovim-prime.mp4
```

## Running the Server

```sh
go run cmd/httpserver/main.go
```

## Endpoints

 - http://localhost:12345/
 - http://localhost:12345/yourproblem
 - http://localhost:12345/myproblem
 - http://localhost:12345/video

### Proxy Endpoints

 - http://localhost:12345/httpbin
 - http://localhost:12345/httpbin/stream/100
 - Etc.
