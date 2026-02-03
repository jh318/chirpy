package main

import (
	"net/http"
	"sync/atomic"
	"fmt"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	mux := http.NewServeMux()
	apiCfg := apiConfig{}
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	wrapped := apiCfg.middlewareMetricsInc(appHandler)
	mux.Handle("/assets/", http.StripPrefix("/assets/",http.FileServer(http.Dir("./assets"))))
	mux.HandleFunc("/healthz", healthzHandler)
	mux.Handle("/app/", wrapped)
	mux.HandleFunc("/metrics", apiCfg.handleMetrics)
	mux.HandleFunc("/reset", apiCfg.handleReset)
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	body := fmt.Sprintf("Hits: %d", hits)
	w.Write([]byte(body))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}