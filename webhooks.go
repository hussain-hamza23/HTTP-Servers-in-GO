package main

import (
	"httpserver/internal/auth"
	"net/http"

	"github.com/google/uuid"
)



func (cfg *apiConfig) upgradeToChirpyRedHandler(w http.ResponseWriter, req *http.Request){
	type chirpyRedPayload struct{
		Event string `json:"event"`
		Data struct{
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	payload, e := decodeRequest[chirpyRedPayload](w, req)
	if e != nil{
		return
	}

	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil || apiKey != string(cfg.polkaKey){
		httpError := &HTTPError{
			StatusCode: http.StatusUnauthorized,
			Message: "Invalid API key",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	if payload.Event != "user.upgraded"{
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := cfg.dbQueries.UpgradeToChirpyRed(req.Context(), payload.Data.UserID); err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error upgrading user to Chirpy Red",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}