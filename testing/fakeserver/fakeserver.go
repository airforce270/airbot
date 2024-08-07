// Package fakeserver provides a fake HTTP server for testing.
package fakeserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// New creates a new FakeServer for testing.
func New() *FakeServer {
	s := FakeServer{}
	s.Reset()
	respIndex := 0
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(fmt.Sprintf("Failed to read request body for URL %s: %v", r.URL.String(), err))
		}

		clonedReq := r.Clone(context.Background())
		clonedReq.Body = io.NopCloser(strings.NewReader(string(body)))
		s.Reqs = append(s.Reqs, clonedReq)

		if len(s.Resps) == 0 {
			return
		}

		// yes, this is hacky
		if strings.Contains(s.Resps[respIndex], `"statusCode": 404`) {
			w.WriteHeader(http.StatusNotFound)
		}

		fmt.Fprint(w, s.Resps[respIndex])

		respIndex++
		if respIndex >= len(s.Resps) {
			respIndex = 0
		}
	}))
	s.s = httpServer
	return &s
}

// FakeServer contains a fake HTTP server for testing.
type FakeServer struct {
	// s is the fake HTTP server.
	s *httptest.Server
	// onClose contains functions to be run when Close() is called.
	onClose []func()

	// Reqs contain the requests made to the server, in order.
	Reqs []*http.Request

	// Resps will be returned in order when calls are made to the server.
	Resps []string
}

// URL returns the server's URL.
func (s *FakeServer) URL(t testing.TB) *url.URL {
	t.Helper()
	u, err := url.Parse(s.s.URL)
	if err != nil {
		t.Fatalf("Failed to parse url %s: %v", s.s.URL, err)
	}
	return u
}

// AddOnClose adds a function to be run when Close() is called.
func (s *FakeServer) AddOnClose(f func()) {
	s.onClose = append(s.onClose, f)
}

// Close closes the server and calls functions registered by AddOnClose()
func (s *FakeServer) Close() {
	for _, f := range s.onClose {
		f()
	}
	s.s.Close()
}

// Reset resets the server's response to its default.
func (s *FakeServer) Reset() {
	s.Resps = nil
}
