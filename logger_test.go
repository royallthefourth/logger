package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type LoggerSuite struct {
	suite.Suite

	req *http.Request
	rl  *responseLogger
	w   *testWriter
}

func (s *LoggerSuite) SetupTest() {
	s.req = httptest.NewRequest(http.MethodGet, "/", nil)
	s.req.RemoteAddr = "192.0.2.1:1234"

	s.rl = &responseLogger{rw: testResponseWriter{}, start: time.Now()}
	s.w = &testWriter{}

	s.rl.Write([]byte("test-logger"))
}

func (s *LoggerSuite) TestRW() {
	s.Equal(s.rl.Header(), s.rl.rw.Header())

	s.rl.WriteHeader(http.StatusAccepted)
	s.Equal(s.rl.status, http.StatusAccepted)
}

func (s *LoggerSuite) TestDefaultHandler() {
	dh := DefaultHandler(http.NotFoundHandler())

	dh.ServeHTTP(s.rl, s.req)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *LoggerSuite) TestSlowHandler() {
	tw := testWriter{}
	dh := Handler(http.NotFoundHandler(), &tw, TinyLogger)

	dh.ServeHTTP(s.rl, s.req)

	s.Equal("GET / 404 19 - 0 ms\n", string(tw.Bytes))
}

func (s *LoggerSuite) TestHandler() {
	tw := testWriter{}
	dh := Handler(http.NotFoundHandler(), &tw, TinyLogger)

	dh.ServeHTTP(s.rl, s.req)

	s.Equal("GET / 404 19 - 0 ms\n", string(tw.Bytes))
}

func (s *LoggerSuite) TestTiny() {
	lh := loggerHandler{
		h:      http.NotFoundHandler(),
		logger: TinyLogger,
		writer: s.w,
	}
	lh.write(s.rl, s.req)

	s.Equal("GET / 200 11 - 0 ms\n", string(s.w.Bytes))
}

func (s *LoggerSuite) TestShort() {
	lh := loggerHandler{
		h:      http.NotFoundHandler(),
		logger: ShortLogger,
		writer: s.w,
	}
	lh.write(s.rl, s.req)

	s.Equal("192.0.2.1:1234 - GET / HTTP/1.1 200 11 - 0 ms\n", string(s.w.Bytes))
}

func (s *LoggerSuite) TestDev() {
	lh := loggerHandler{
		h:      http.NotFoundHandler(),
		logger: DevLogger,
		writer: s.w,
	}
	lh.write(s.rl, s.req)

	s.Equal("GET / 200 0 ms - 11\n", string(s.w.Bytes))
}

func (s *LoggerSuite) TestCommon() {
	lh := loggerHandler{
		h:      http.NotFoundHandler(),
		logger: CommonLogger,
		writer: s.w,
	}
	lh.write(s.rl, s.req)

	s.Equal(`192.0.2.1:1234 - - [`+s.rl.start.Format(timeFormat)+`] "GET / HTTP/1.1" 200 11`+"\n", string(s.w.Bytes))
}

func (s *LoggerSuite) TestCombined() {
	lh := loggerHandler{
		h:      http.NotFoundHandler(),
		logger: CombinedLogger,
		writer: s.w,
	}
	lh.write(s.rl, s.req)

	s.Equal(`192.0.2.1:1234 - - [`+s.rl.start.Format(timeFormat)+`] "GET / HTTP/1.1" 200 11 "" ""`+"\n", string(s.w.Bytes))
}

func TestLogger(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}

type testResponseWriter struct {
	header http.Header
}

func (trw testResponseWriter) Header() http.Header {
	if trw.header != nil {
		return trw.header
	}

	return make(http.Header)
}

func (trw testResponseWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}

func (trw testResponseWriter) WriteHeader(status int) {}

type testWriter struct {
	Bytes []byte
}

func (tw *testWriter) Write(b []byte) (n int, err error) {
	tw.Bytes = append(tw.Bytes, b...)

	return len(b), nil
}

func Test_parseResponseTime(t *testing.T) {
	start := time.Now().Add(time.Duration(-10) * time.Millisecond)
	t.Logf("Time diff of %s\n", parseResponseTime(start))
}
