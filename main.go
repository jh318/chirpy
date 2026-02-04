package main

import (
	"net/http"
	"sync/atomic"
	"fmt"
	"encoding/json"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	mux := http.NewServeMux()
	apiCfg := apiConfig{}
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	wrapped := apiCfg.middlewareMetricsInc(appHandler)
	mux.Handle("/app/", wrapped)
	mux.HandleFunc("GET /admin/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handleReset)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handleValidateChirp)
	server := http.Server{}
	server.Handler = mux
	server.Addr = ":8080"
	server.ListenAndServe()
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})

}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	hits := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	body := fmt.Sprintf(
		`<html>
  			<body>
    			<h1>Welcome, Chirpy Admin</h1>
    			<p>Chirpy has been visited %d times!</p>
  			</body>
		</html>`, hits)
	w.Write([]byte(body))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type successParameters struct {
		Valid bool `json:"valid"`
	}
	type errorParameters struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		errResp := errorParameters{Error: "Something went wrong"}
		data, _ := json.Marshal(errResp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(data)
		return
	} else if len(params.Body) > 140 {
		errResp := errorParameters{Error: "Chirp is too long"}
		data, _ := json.Marshal(errResp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}

	p := successParameters{Valid: true}
	data, _ := json.Marshal(p)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}