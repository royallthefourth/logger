package logger_test

import (
	"net/http"
	"os"

	"github.com/royallthefourth/logger"
)

func Example_defaultHandler() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("Hello World"))
	})

	http.ListenAndServe(":8080", logger.DefaultHandler(mux))
}

func Example_handler() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("Hello World"))
	})

	http.ListenAndServe(":8080", logger.Handler(mux, os.Stdout,
		logger.DevLogger))
}
