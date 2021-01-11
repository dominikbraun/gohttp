# gohttp

`gohttp` provides convenience functions for parsing `net/http` objects from a source and
serializing them back to byte slices. At the moment, `gohttp` is HTTP/1.1-compatible only.

**Status:** WIP and experimental. If you intend to parse requests from a byte source, use
[`http.ReadRequest`](https://golang.org/pkg/net/http/#ReadRequest) instead.

## Examples

### ParseRequest

ParseRequest reads a given source and parses an `http.Request` instance from it.

```go
conn, err := server.listener.Accept()

// Handle errors, defer closing the connection etc.

conn.setReadDeadline(time.Now().Add(10 * time.Second))
connReader := bufio.NewReader(conn)

request, _ := gohttp.ParseRequest(connReader)
```

### SerializeRequest

SerializeRequest converts an `http.Request` instance into a byte slice.

```go
conn.setWriteDeadline(time.Now().Add(10 * time.Second))
serializedRequest, _ := gohttp.SerializeRequest(request)
_, _ = conn.Write(serializedRequest)
```
