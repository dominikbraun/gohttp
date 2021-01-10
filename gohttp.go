// Package gohttp provides convenience functions for parsing net/http
// objects from a source and converting them back to byte slices.
package gohttp

import (
	"io"
	"net/http"
)

// ParseRequest reads a given source and parses an http.Request instance
// from it. In case an error occurs, a zero-value Request is returned.
//
// The HTTP request has to be formatted according to RFC 7230, section 3.
// If the user allows LF line endings, the header fields and the empty
// line terminating the header section may be LF instead of CRLF endings.
//
// When using ParseRequest to parse data from a net.Conn object, make sure
// to set an appropriate read deadline on the net.Conn instance first.
func ParseRequest(r io.Reader) (*http.Request, error) {
	return nil, nil
}

// SerializeRequest converts an http.Request instance into a byte slice
// according to the HTTP message syntax defined in RFC 7230, section 3.
//
// SerializeRequest uses CRLF line endings when serializing the request
// instance, regardless whether the user allows LF line endings or not.
func SerializeRequest(r *http.Request) (error, []byte) {
	return nil, nil
}

// ParseResponse reads a given source and parses an http.Response instance
// form it. In case an error occurs, a zero-value Response is returned.
//
// The HTTP response has to be formatted according to RFC 7230, section 3.
// If the user allows LF line endings, the header fields and the empty
// line terminating the header section may be LF instead of CRLF endings.
func ParseResponse(r io.Reader) (*http.Response, error) {
	return nil, nil
}

// SerializeResponse converts an http.Response instance into a byte slice
// according to the HTTP message syntax defined in RFC 7230, section 3.
//
// SerializeResponse uses CRLF line endings when serializing the response
// instance, regardless whether the user allows LF line endings or not.
func SerializeResponse(r *http.Response) (error, []byte) {
	return nil, nil
}
