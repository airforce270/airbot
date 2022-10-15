// Package fakeserver provides a fake HTTP server for testing.
package fakeserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func New() *FakeServer {
	s := FakeServer{}
	s.Reset()
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, s.Resp)
	}))
	s.s = httpServer
	return &s
}

type FakeServer struct {
	s       *httptest.Server
	onClose []func()

	Resp string
}

func (s *FakeServer) URL() string {
	return s.s.URL
}

func (s *FakeServer) AddOnClose(f func()) {
	s.onClose = append(s.onClose, f)
}

func (s *FakeServer) Close() {
	for _, f := range s.onClose {
		f()
	}
	s.s.Close()
}

func (s *FakeServer) Reset() {
	s.Resp = "no-response-set"
}
