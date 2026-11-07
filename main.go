package main

import (
	"net/http"
	"log"
)

func main() {
	const port = "8080"
	
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	mux.Handle("/", fileServer)
	server := &http.Server {
		Handler: mux,
		Addr: port,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())


}
	