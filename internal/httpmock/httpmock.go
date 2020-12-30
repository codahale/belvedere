package httpmock

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type expectation struct {
	url      url.URL
	method   string
	req      string
	status   int
	resp     string
	optional bool
	called   bool
}

type Server struct {
	m            sync.Mutex
	srv          *httptest.Server
	t            testing.TB
	expectations []expectation
}

type Option func(e *expectation)

func NewServer(t testing.TB) *Server {
	t.Helper()

	server := &Server{t: t}
	server.srv = httptest.NewServer(http.HandlerFunc(server.handle))

	return server
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	s.t.Helper()
	s.m.Lock()
	defer s.m.Unlock()

	for i, exp := range s.expectations {
		if !exp.called && *r.URL == exp.url {
			s.expectations[i].called = true
			s.checkExpectation(w, r, exp)

			return
		}
	}

	http.NotFound(w, r)
	s.t.Errorf("Unexpected request for %q", r.URL.String())
}

func (s *Server) checkExpectation(w http.ResponseWriter, r *http.Request, exp expectation) {
	if exp.method != "" {
		assert.Equal(s.t, "method", exp.method, r.Method,
			cmpopts.AcyclicTransformer("ToUpper", strings.ToUpper))
	}

	if exp.req != "" {
		req, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.t.Fatal(err)
		}

		_ = r.Body.Close()

		assert.Equal(s.t, "request", exp.req, string(req),
			cmpopts.AcyclicTransformer("TrimSpace", strings.TrimSpace))
	}

	if exp.status != 0 {
		w.WriteHeader(exp.status)
	}

	if exp.resp != "" {
		_, err := io.WriteString(w, exp.resp)
		if err != nil {
			s.t.Fatal(err)
		}
	}
}

func (s *Server) URL() string {
	s.t.Helper()
	return s.srv.URL
}

func (s *Server) Client() *http.Client {
	s.t.Helper()
	return s.srv.Client()
}

func (s *Server) Expect(reqURL string, opt ...Option) {
	s.t.Helper()
	s.m.Lock()
	defer s.m.Unlock()

	u, err := url.Parse(reqURL)
	if err != nil {
		s.t.Fatal(err)
	}

	e := expectation{
		url: *u,
	}

	for _, f := range opt {
		f(&e)
	}

	s.expectations = append(s.expectations, e)
}

func RespJSON(resp interface{}) Option {
	j, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	return func(e *expectation) {
		e.resp = string(j)
	}
}

func Method(method string) Option {
	return func(e *expectation) {
		e.method = method
	}
}

func Status(status int) Option {
	return func(e *expectation) {
		e.status = status
	}
}

func ReqJSON(req interface{}) Option {
	j, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	return func(e *expectation) {
		e.req = string(j)
	}
}

func Optional() Option {
	return func(e *expectation) {
		e.optional = true
	}
}

func (s *Server) Finish() {
	s.t.Helper()
	s.m.Lock()
	defer s.m.Unlock()

	for _, exp := range s.expectations {
		if !exp.optional && !exp.called {
			s.t.Errorf("No request for %q", exp.url.String())
		}
	}
}
