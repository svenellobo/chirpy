package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const port = "8080"
	
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	var apiCfg apiConfig

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/metrics", apiCfg.handlerRequestCounter)
	mux.HandleFunc("POST /api/reset", apiCfg.handlerResetCounter)

	
	server := &http.Server {
		Handler: mux,
		Addr: ":" + port,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())


}

func handlerReadiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")	
	w.WriteHeader(200)
	w.Write([]byte("OK\n"))
	
	}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w,r)
	})
}

func (cfg *apiConfig) handlerRequestCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	reqNumb := cfg.fileserverHits.Load()
	w.WriteHeader(http.StatusOK)
	text := fmt.Sprintf("Hits: %d\n", reqNumb)
	w.Write([]byte(text))
	
}

func (cfg *apiConfig) handlerResetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request counter has been reset\n"))
}

