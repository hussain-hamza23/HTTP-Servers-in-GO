package main

import (
	"encoding/json"
	"httpserver/internal/auth"
	"net/http"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, req *http.Request){
	type requestPayload struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}

	var payload requestPayload = requestPayload{}
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := decoder.Decode(&payload); err != nil{
		errorResponse(w, http.StatusInternalServerError, "Error decoding parameters for user login")
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(req.Context(), payload.Email)
	if err != nil{
		errorResponse(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	match, err := auth.CheckPasswordHash(payload.Password, user.Password)
	if err != nil{
		errorResponse(w, http.StatusInternalServerError, "Error checking password")
		return
	}
	if !match {
		errorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	var userResponse User = User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}

	jsonResponse(w, http.StatusOK, userResponse)
}