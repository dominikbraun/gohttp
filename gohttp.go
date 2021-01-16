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
	"unicode"
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
func ParseRequest(reader *bufio.Reader) (*http.Request, error) {
	request := http.Request{}

	// RFC 7230, section 3.5. states that a robust parser implementation
	// should ignore at least one empty line prior to the request line.
	for {
		line, err := reader.ReadString('\n')
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

	var line string
	var err error
	for {
		line, err = reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if isNewLine(line) {
			break
		}

		fieldName, fieldValue, err := parseHeaderField(line)
		if err != nil {
			return nil, err
		}

		request.Header.Add(fieldName, fieldValue)
	}

	if !isNewLine(line) {
		return nil, errors.New("empty line after header section is missing")
	}

	length, err := determineBodyLength(request.Header, reader)
	if err != nil {
		return nil, err
	}

	if length > 0 {
		var body = make([]byte, length)
		n, err := reader.Read(body)
		if err != nil {
			return nil, err
		}
		if n != length {
			return nil, errors.New("wrong body length")
		}
	}

	return &request, nil
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
func ParseResponse(reader *bufio.Reader) (*http.Response, error) {
	response := http.Response{}

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	protocol, statusCode, reasonPhrase, err := parseStatusLine(line)
	if err != nil {
		return nil, err
	}

	response.Proto = protocol
	response.StatusCode = statusCode
	response.Status = fmt.Sprintf("%d %s", statusCode, reasonPhrase)

	response.Header = make(http.Header)

	for {
		line, err = reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if isNewLine(line) {
			break
		}

		fieldName, fieldValue, err := parseHeaderField(line)
		if err != nil {
			return nil, err
		}

		response.Header.Add(fieldName, fieldValue)
	}

	if !isNewLine(line) {
		return nil, errors.New("empty line after header section is missing")
	}

	length, err := determineBodyLength(response.Header, reader)
	if err != nil {
		return nil, err
	}

	if length > 0 {
		var body = make([]byte, length)
		n, err := reader.Read(body)
		if err != nil {
			return nil, err
		}
		if n != length {
			return nil, errors.New("wrong body length")
		}
	}

	return &response, nil
}

// SerializeResponse converts an http.Response instance into a byte slice.
//
// SerializeResponse uses CRLF line endings when serializing the response
// instance, regardless whether the user allows LF line endings or not.
func SerializeResponse(r *http.Response) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s %s\r\n", r.Proto, r.Status))

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

func parseRequestLine(line string) (string, *url.URL, string, error) {
	data := strings.Split(line, " ")

	// RFC 7230, section 3.1.1. prescribes exactly 3 tokens.
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

func parseStatusLine(line string) (string, int, string, error) {
	data := strings.Split(line, " ")

	// RFC 7230, section 3.1.2. prescribes exactly 3 tokens.
	if len(data) != 3 {
		return "", 0, "", errors.New("invalid status line syntax")
	}

	protocol := strings.TrimSuffix(data[0], "\n")
	statusCode := strings.TrimSuffix(data[1], "\n")
	reasonPhrase := strings.TrimSuffix(data[2], "\n")

	parsedStatusCode, err := strconv.Atoi(statusCode)
	if err != nil {
		return "", 0, "", err
	}

	return protocol, parsedStatusCode, reasonPhrase, nil
}

func parseHeaderField(line string) (string, string, error) {
	tokens := strings.SplitN(line, ":", 2)

	// RFC 7230, sections 3.2. and 3.2.4. prescribe exactly 2 tokens.
	if len(tokens) != 2 {
		return "", "", errors.New("invalid header field syntax")
	}

	name := strings.TrimSpace(tokens[0])
	value := strings.TrimSpace(strings.TrimSuffix(tokens[1], "\n"))

	return name, value, nil
}

func writeHeaderFields(headers http.Header, w io.Writer) error {
	for fieldName, values := range headers {
		var fieldValue string

		// Assemble the field value components (RFC 7230, section 3.2.6).
		for i := 0; i < len(values); i++ {
			fieldValue += values[i]
			if i < len(values)-1 {
				fieldValue += ", "
			}
		}

		headerField := fmt.Sprintf("%s: %s\r\n", fieldName, fieldValue)

		_, err := w.Write([]byte(headerField))
		if err != nil {
			return err
		}
	}

	return nil
}

func determineBodyLength(headers http.Header, reader *bufio.Reader) (int, error) {

	// If the Transfer-Encoding header is set, the length of the message
	// chunk is contained within the body (RFC 7230, section 3.3.3.).
	if transferEncoding := headers.Get("Transfer-Encoding"); transferEncoding != "" {
		firstBodyLine, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}

		// Parse the hex code as int.
		// Use TrimRightFunc and not strings.TrimSpace to make sure the hex is at the beginning.
		length, err := strconv.ParseInt(strings.TrimRightFunc(firstBodyLine, unicode.IsSpace), 16, 64)
		if err != nil {
			return 0, err
		}

		return int(length), nil
	}

	if contentLength := headers.Get("Content-Length"); contentLength != "" {
		return strconv.Atoi(contentLength)
	}

	return 0, nil
}

func isNewLine(line string) bool {
	if allowLFLineEndings {
		return line == "\r\n" || line == "\n"
	}
	return line == "\r\n"
}
