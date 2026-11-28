package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

var forbiddenWords = map[string]struct{} {
		"kerfuffle": {},
		"sharbert": {},
		"fornax": {},
	}

func main() {
	const port = "8080"
	
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	var apiCfg apiConfig

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerRequestCounter)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetCounter)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	reqNumb := cfg.fileserverHits.Load()
	w.WriteHeader(http.StatusOK)
	 
	text := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, reqNumb)
	w.Write([]byte(text))
	
}

func (cfg *apiConfig) handlerResetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request counter has been reset\n"))
}

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type parameter struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameter {}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Something went wrong")		
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	

	type cleanedBody struct {
		Clean string `json:"cleaned_body"`
	}

	clean := cleanedBody {
		Clean: replaceForbiddenWord(params.Body),
	}
	
	respondWithJSON(w, 200, clean)

}

func replaceForbiddenWord(body string) string {
	words := strings.Split(body, " ")
	
	for i, word := range words{
		if _,ok := forbiddenWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")

}
