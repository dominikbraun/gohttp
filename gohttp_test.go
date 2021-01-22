package gohttp

import (
	"bufio"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// TestParseRequest tests the ParseRequest function by providing sample
// requests with different HTTP methods. The parsed http.Request instance
// is expected to possess the same data as the message struct.
func TestParseRequest(t *testing.T) {
	type message struct {
		method   string
		url      string
		protocol string
		body     string
	}

	testCases := map[string]struct {
		source   string
		expected message
	}{
		"GET request": {
			source: `
GET / HTTP/1.1
Host: www.example.com

`,
			expected: message{
				method:   "GET",
				url:      "/",
				protocol: "HTTP/1.1",
			},
		},
	}

	for name, tc := range testCases {
		reader := bufio.NewReader(strings.NewReader(tc.source))
		actual, err := ParseRequest(reader, WithLFLineEndings(true))
		if err != nil {
			t.Fatalf("'%s': unexpected error: %s", name, err.Error())
		}

		if actual.URL.String() != tc.expected.url {
			t.Errorf("'%s': expected method %s, got %s", name, tc.expected.url, actual.URL.String())
		}

		if actual.Method != tc.expected.method {
			t.Errorf("'%s': expected target URL %s, got %s", name, tc.expected.method, actual.Method)
		}

		if actual.Proto != tc.expected.protocol {
			t.Errorf("'%s': expected protocol %s, got %s", name, tc.expected.protocol, actual.Proto)
		}
	}
}

// TestSerializeRequest tests the SerializeRequest function by providing
// sample http.Request instances with different HTTP methods. The serialized
// request is expected to be identical to the string representation.
func TestSerializeRequest(t *testing.T) {
	parsedUrl, _ := url.Parse("/")

	testCases := map[string]struct {
		request  *http.Request
		expected string
	}{
		"GET request": {
			request: &http.Request{
				Method:     "GET",
				URL:        parsedUrl,
				Proto:      "HTTP/1.1",
				ProtoMajor: 0,
				ProtoMinor: 0,
				Header: map[string][]string{
					"Host":              {"example.com"},
					"Transfer-Encoding": {"gzip", "chunked"},
				},
				Body: http.NoBody,
			},
			expected: "GET / HTTP/1.1\r\n" +
				"Host: example.com\r\n" +
				"Transfer-Encoding: gzip, chunked\r\n" +
				"\r\n",
		},
	}

	for name, tc := range testCases {
		actual, err := SerializeRequest(tc.request)
		if err != nil {
			t.Fatalf("'%s': unexpected error: %s", name, err.Error())
		}

		if string(actual) != tc.expected {
			t.Errorf("'%s': expected request %s, got %s", name, tc.expected, string(actual))
		}
	}
}

func TestParseResponse(t *testing.T) {}

func TestSerializeResponse(t *testing.T) {}

func TestParseRequestLine(t *testing.T) {
	type requestLine struct {
		method    string
		parsedURL string
		protocol  string
	}

	testCases := map[string]struct {
		line     string
		expected requestLine
	}{
		"GET request": {
			line: "GET example.com HTTP/1.1",
			expected: requestLine{
				method:    "GET",
				parsedURL: "example.com",
				protocol:  "HTTP/1.1",
			},
		},
	}

	for name, tc := range testCases {
		actualMethod, actualURL, actualProtocol, err := parseRequestLine(tc.line)
		if err != nil {
			t.Fatalf("'%s': unexpected error: %s", name, err.Error())
		}

		if actualMethod != tc.expected.method {
			t.Errorf("'%s': expected method %s, got %s", name, tc.expected.method, actualMethod)
		}

		if actualURL.String() != tc.expected.parsedURL {
			t.Errorf("'%s': expected URL %s, got %s", name, tc.expected.method, actualMethod)
		}

		if actualProtocol != tc.expected.protocol {
			t.Errorf("'%s': expected protocol %s, got %s", name, tc.expected.method, actualMethod)
		}
	}
}

func TestParseStatusLine(t *testing.T) {
	type statusLine struct {
		protocol     string
		statusCode   int
		reasonPhrase string
	}

	testCases := map[string]struct {
		line     string
		expected statusLine
	}{
		"GET request": {
			line: "HTTP/1.1 200 OK",
			expected: statusLine{
				protocol:     "HTTP/1.1",
				statusCode:   200,
				reasonPhrase: "OK",
			},
		},
	}

	for name, tc := range testCases {
		actualProtocol, actualStatusCode, actualReasonPhrase, err := parseStatusLine(tc.line)
		if err != nil {
			t.Fatalf("'%s': unexpected error: %s", name, err.Error())
		}

		if actualProtocol != tc.expected.protocol {
			t.Errorf("'%s': expected protocol %s, got %s", name, tc.expected.protocol, actualProtocol)
		}

		if actualStatusCode != tc.expected.statusCode {
			t.Errorf("'%s': expected status code %d, got %d", name, tc.expected.statusCode, actualStatusCode)
		}

		if actualReasonPhrase != tc.expected.reasonPhrase {
			t.Errorf("'%s': expected reason phrase %s, got %s", name, tc.expected.reasonPhrase, actualReasonPhrase)
		}
	}
}

func TestParseHeaderField(t *testing.T) {
	type headerField struct {
		name  string
		value string
	}

	testCases := map[string]struct {
		line     string
		expected headerField
	}{
		"Content-Length header": {
			line: "Content-Length: 1024",
			expected: headerField{
				name:  "Content-Length",
				value: "1024",
			},
		},
	}

	for name, tc := range testCases {
		actualName, actualValue, err := parseHeaderField(tc.line)
		if err != nil {
			t.Fatalf("'%s': unexpected error: %s", name, err.Error())
		}

		if actualName != tc.expected.name {
			t.Errorf("'%s': expected name %s, got %s", name, tc.expected.name, actualName)
		}

		if actualValue != tc.expected.value {
			t.Errorf("'%s': expected value %s, got %s", name, tc.expected.value, actualValue)
		}
	}
}

func TestWriteHeaderFields(t *testing.T) {
	return
}

func TestDetermineBodyLength(t *testing.T) {
	return
}

func TestIsNewLine(t *testing.T) {
	return
}
