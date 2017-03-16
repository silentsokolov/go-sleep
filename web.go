package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/silentsokolov/go-sleep/log"
)

var templateFuncMap = template.FuncMap{
	"CheckExistsTime": func(i *time.Time) bool {
		if i == nil {
			return false
		}
		return true
	},
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func startWebServer(addr string) {
	srv := http.NewServeMux()

	srv.HandleFunc("/", indexHandler)

	log.Printf("Starting web server on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal("Error creating web server: ", err)
	}
}
