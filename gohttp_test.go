package gohttp

import (
	"bufio"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type message struct {
	method   string
	url      string
	protocol string
	body     string
}

// TestParseRequest tests the ParseRequest function by providing sample
// requests with different HTTP methods. The parsed http.Request instance
// is expected to possess the same data as the message struct.
func TestParseRequest(t *testing.T) {
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
				"Transfer-Encoding: gzip, chunked\r\n",
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
