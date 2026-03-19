package main

import "net/http"

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request){
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte("Hits reset to 0")); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}