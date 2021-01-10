// Package gohttp provides convenience functions for parsing net/http
// objects from a source and converting them back to byte slices.
package gohttp

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	// allowLFLineEndings determines whether LF line endings are valid
	// for situations where RFC 7230 prescribes CRLF line endings.
	allowLFLineEndings = false
)

// AllowLFLineEndings defines whether LF line endings are allowed as a
// replacement for CRLF line endings for a more tolerant source parsing.
//
// AllowLFLineEndings doesn't have any effect on serialization functions.
func AllowLFLineEndings(allow bool) {
	allowLFLineEndings = allow
}

// ParseRequest reads a given source and parses an http.Request instance
// from it. In case an error occurs, a zero-value Request is returned.
//
// The HTTP request has to be formatted according to RFC 7230, section 3.
// If the user allows LF line endings, the header fields and the empty
// line terminating the header section may be LF instead of CRLF endings.
//
// When using ParseRequest to parse data from a net.Conn object, make sure
// to set an appropriate read deadline on the net.Conn instance first.
func ParseRequest(r bufio.Reader) (*http.Request, error) {
	request := http.Request{}

	// RFC 7230, section 3.5. states that a robust parser implementation
	// should ignore at least one empty line prior to the request line.
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if !isNewLine(line) {
			if err := readInRequestLine(line, &request); err != nil {
				return nil, err
			}
			break
		}
	}

	for {
		line, err := r.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		if errors.Is(err, io.EOF) || isNewLine(line) {
			break
		}

		fieldName, fieldValue, err := parseHeaderField(line)
		if err != nil {
			return nil, err
		}

		request.Header.Add(fieldName, fieldValue)
	}

	if length := determineBodyLength(&request); length > 0 {
		_, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		// ToDo: Set request body
	}

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
func ParseResponse(r bufio.Reader) (*http.Response, error) {
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

// readInRequestLine reads in an HTTP request line and populates a given
// request object with the parsed data.
//
// The line has to be formatted according to RFC 7230, section 3.1.1.
func readInRequestLine(requestLine string, r *http.Request) error {
	data := strings.Split(requestLine, " ")

	if len(data) != 3 {
		return errors.New("invalid request line syntax")
	}

	targetUrl, err := url.Parse(data[1])
	if err != nil {
		return err
	}

	r.Method = data[0]
	r.Proto = data[2]
	r.URL = targetUrl

	return nil
}

// parseHeaderField parses a single HTTP header field. The respective line
// has to be formatted according to RFC 7230, sections 3.2. and 3.2.4.
func parseHeaderField(line string) (string, string, error) {
	tokens := strings.Split(line, ":")

	if len(tokens) != 2 {
		return "", "", errors.New("invalid header field syntax")
	}

	name := strings.TrimSpace(tokens[0])
	value := strings.TrimSpace(tokens[1])

	return name, value, nil
}

// determineBodyLength determines the expected body length for a given
// HTTP request as described in FRC 7230, section 3.3.
func determineBodyLength(r *http.Request) int {
	if transferEncoding := r.Header.Get("Transfer-Encoding"); transferEncoding != "" {
		// ToDo: Determine the body length based on Transfer-Encoding
		return 1
	}

	if contentLength := r.Header.Get("Content-Length"); contentLength != "" {
		length, _ := strconv.Atoi(contentLength)
		return length
	}

	return 0
}

func isNewLine(line string) bool {
	if allowLFLineEndings {
		return line == "\r\n" || line == "\n"
	}
	return line == "\n"
}
