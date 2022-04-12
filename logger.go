package logger

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const timeFormat = "02/Jan/2006:15:04:05 -0700"

type Logger func(rl *responseLogger, req *http.Request, username string) string

// CombinedLogger is the standard Apache combined log output
//
// format:
//
// :remote-addr - :remote-user [:date[clf]] ":method :url
// HTTP/:http-version" :status :res[content-length] ":referrer" ":user-agent"
func CombinedLogger(rl *responseLogger, req *http.Request, username string) string {
	return strings.Join([]string{
		req.RemoteAddr,
		"-",
		username,
		"[" + rl.start.Format(timeFormat) + "]",
		`"` + req.Method,
		req.RequestURI,
		req.Proto + `"`,
		strconv.Itoa(rl.status),
		strconv.Itoa(rl.size),
		`"` + req.Referer() + `"`,
		`"` + req.UserAgent() + `"`,
	}, " ")
}

// CommonLogger is the standard Apache common log output
//
// format:
//
// :remote-addr - :remote-user [:date[clf]] ":method :url
// HTTP/:http-version" :status :res[content-length]
func CommonLogger(rl *responseLogger, req *http.Request, username string) string {
	return strings.Join([]string{
		req.RemoteAddr,
		"-",
		username,
		"[" + rl.start.Format(timeFormat) + "]",
		`"` + req.Method,
		req.RequestURI,
		req.Proto + `"`,
		strconv.Itoa(rl.status),
		strconv.Itoa(rl.size),
	}, " ")
}

// DevLogger is useful for development
//
// format:
//
// :method :url :status :response-time ms - :res[content-length]
func DevLogger(rl *responseLogger, req *http.Request, username string) string {
	return strings.Join([]string{
		req.Method,
		req.RequestURI,
		strconv.Itoa(rl.status),
		parseResponseTime(rl.start),
		"-",
		strconv.Itoa(rl.size),
	}, " ")
}

// ShortLogger is shorter than common, but includes response time
//
// format:
//
// :remote-addr :remote-user :method :url HTTP/:http-version :status
// :res[content-length] - :response-time ms
func ShortLogger(rl *responseLogger, req *http.Request, username string) string {
	return strings.Join([]string{
		req.RemoteAddr,
		username,
		req.Method,
		req.RequestURI,
		req.Proto,
		strconv.Itoa(rl.status),
		strconv.Itoa(rl.size),
		"-",
		parseResponseTime(rl.start),
	}, " ")
}

// TinyLogger is the smallest format
//
// format:
//
// :method :url :status :res[content-length] - :response-time ms
func TinyLogger(rl *responseLogger, req *http.Request, username string) string {
	return strings.Join([]string{
		req.Method,
		req.RequestURI,
		strconv.Itoa(rl.status),
		strconv.Itoa(rl.size),
		"-",
		parseResponseTime(rl.start),
	}, " ")
}

type responseLogger struct {
	rw     http.ResponseWriter
	start  time.Time
	status int
	size   int
}

func (rl *responseLogger) Header() http.Header {
	return rl.rw.Header()
}

func (rl *responseLogger) Write(bytes []byte) (int, error) {
	if rl.status == 0 {
		rl.status = http.StatusOK
	}

	size, err := rl.rw.Write(bytes)

	rl.size += size

	return size, err
}

func (rl *responseLogger) WriteHeader(status int) {
	rl.status = status

	rl.rw.WriteHeader(status)
}

func (rl *responseLogger) Flush() {
	f, ok := rl.rw.(http.Flusher)

	if ok {
		f.Flush()
	}
}

type loggerHandler struct {
	logger Logger
	h      http.Handler
	writer io.Writer
}

func (rh loggerHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	rl := &responseLogger{rw: res, start: time.Now()}

	rh.h.ServeHTTP(rl, req)

	rh.write(rl, req)
}

func (rh loggerHandler) write(rl *responseLogger, req *http.Request) {
	username := "-"

	if req.URL.User != nil {
		if name := req.URL.User.Username(); name != "" {
			username = name
		}
	}

	fmt.Fprintln(rh.writer, rh.logger(rl, req, username))
}

func parseResponseTime(start time.Time) string {
	return fmt.Sprintf("%d ms", time.Now().Sub(start).Milliseconds())
}

// DefaultHandler returns a http.Handler that wraps h by using
// Apache combined log output and prints to os.Stdout
func DefaultHandler(h http.Handler) http.Handler {
	return loggerHandler{
		h:      h,
		logger: CombinedLogger,
		writer: os.Stdout,
	}
}

// Handler returns a http.Handler that wraps h by using t type log output
// and print to writer
func Handler(h http.Handler, writer io.Writer, logger Logger) http.Handler {
	return loggerHandler{
		h:      h,
		logger: logger,
		writer: writer,
	}
}
