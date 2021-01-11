// Package gohttp provides convenience functions for parsing net/http
// objects from a source and serializing them back to byte slices.
package gohttp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	allowLFLineEndings = false
)

// AllowLFLineEndings defines whether LF line endings are allowed for more
// tolerant parsing. Use with care!
func AllowLFLineEndings(allow bool) {
	allowLFLineEndings = allow
}

// ParseRequest reads a given source and parses an http.Request instance
// from it.
//
// If the user allows LF line endings, the header fields and the empty
// line terminating the header section may be LF instead of CRLF endings.
func ParseRequest(r *bufio.Reader) (*http.Request, error) {
	request := http.Request{}

	// RFC 7230, section 3.5. states that a robust parser implementation
	// should ignore at least one empty line prior to the request line.
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if !isNewLine(line) {
			method, targetUrl, protocol, err := parseRequestLine(line)
			if err != nil {
				return nil, err
			}

			request.Method = method
			request.URL = targetUrl
			request.Proto = protocol

			break
		}
	}

	request.Header = make(http.Header)

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

	emptyLine, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	// If the error is EOF, a body isn't expected anymore. Otherwise, an
	// empty line followed by the actual message body is expected.
	if errors.Is(err, io.EOF) {
		return &request, nil
	}

	if !isNewLine(emptyLine) {
		return nil, errors.New("empty line after header section is missing")
	}

	bodyOnset, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return &request, nil
		}
		return nil, err
	}

	if length := determineBodyLength(request.Header, bodyOnset); length > 0 {
		// ToDo: Write remaining message into request.Body
	}

	return nil, nil
}

// SerializeRequest converts an http.Request instance into a byte slice.
//
// SerializeRequest uses CRLF line endings when serializing the request
// instance, regardless whether the user allows LF line endings or not.
func SerializeRequest(r *http.Request) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s %s %s\r\n", r.Method, r.URL.String(), r.Proto))

	for fieldName, values := range r.Header {
		var fieldValue string

		// Assemble the field value components (RFC 7230, section 3.2.6).
		for i := 0; i < len(values); i++ {
			fieldValue += values[i]
			if i < len(values)-1 {
				fieldValue += ", "
			}
		}

		buf.WriteString(fmt.Sprintf("%s: %s\r\n", fieldName, fieldValue))
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return buf.Bytes(), nil
	}

	buf.WriteString("\r\n")
	buf.Write(body)

	return buf.Bytes(), nil
}

// ParseResponse reads a given source and parses an http.Response instance
// from it.
//
// If the user allows LF line endings, the header fields and the empty
// line terminating the header section may be LF instead of CRLF endings.
func ParseResponse(r *bufio.Reader) (*http.Response, error) {
	return nil, nil
}

// SerializeResponse converts an http.Response instance into a byte slice.
//
// SerializeResponse uses CRLF line endings when serializing the response
// instance, regardless whether the user allows LF line endings or not.
func SerializeResponse(r *http.Response) ([]byte, error) {
	return nil, nil
}

func parseRequestLine(line string) (string, *url.URL, string, error) {
	data := strings.Split(line, " ")

	// RFC 7230, section 3.1.1 prescribes exactly 3 tokens.
	if len(data) != 3 {
		return "", nil, "", errors.New("invalid request line syntax")
	}

	method := strings.TrimSuffix(data[0], "\n")
	targetUrl := strings.TrimSuffix(data[1], "\n")
	protocol := strings.TrimSuffix(data[2], "\n")

	parsedUrl, err := url.Parse(targetUrl)
	if err != nil {
		return "", nil, "", err
	}

	return method, parsedUrl, protocol, nil
}

func parseHeaderField(line string) (string, string, error) {
	tokens := strings.Split(line, ":")

	// RFC 7230, sections 3.2. and 3.2.4. prescribe exactly 2 tokens.
	if len(tokens) != 2 {
		return "", "", errors.New("invalid header field syntax")
	}

	name := strings.TrimSpace(tokens[0])
	value := strings.TrimSpace(strings.TrimSuffix(tokens[1], "\n"))

	return name, value, nil
}

func determineBodyLength(headers http.Header, bodyOnset string) int {

	// If the Transfer-Encoding header is set, the length of the message
	// chunk is contained within the body (RFC 7230, section 3.3.3.).
	if transferEncoding := headers.Get("Transfer-Encoding"); transferEncoding != "" {
		// ToDo: Determine the body length based on Transfer-Encoding
		return 1
	}

	if contentLength := headers.Get("Content-Length"); contentLength != "" {
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
