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

func (cfg *apiConfig) resetUsersHandler(w http.ResponseWriter, r *http.Request){
	if cfg.role != "dev"{
		errorResponse(w, http.StatusForbidden, "Forbidden: insufficient permissions")
		return
	}
	if err := cfg.dbQueries.DeleteUsers(r.Context()); err != nil{
		errorResponse(w, http.StatusInternalServerError, "Error resetting users")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte("Users reset successfully")); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}