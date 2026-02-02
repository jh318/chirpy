package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.Handle("/assets/", http.StripPrefix("/assets/",http.FileServer(http.Dir("./assets"))))
	server := http.Server{}
	server.Handler = mux
	server.Addr = ":8080"
	server.ListenAndServe()
}
