package main

import (
	"log"
	"net/http"

	"github.com/Reece-Reklai/go_serve/internal"
)

func main() {
	router := internal.Router{Mux: http.NewServeMux()}
	port := "8080"
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router.Mux,
	}
	staticDir := "./public/"
	router.Mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	log.Fatal(server.ListenAndServe())
}
