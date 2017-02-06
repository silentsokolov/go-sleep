package main

import (
	"fmt"
	"log"
	"net/http"
)

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