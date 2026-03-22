package main

import (
	"encoding/json"
	"httpserver/internal/auth"
	"httpserver/internal/database"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type User struct{
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, req *http.Request){
	type requestPayload struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}

	var payload requestPayload = requestPayload{}
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := decoder.Decode(&payload); err != nil{
		errorResponse(w, http.StatusInternalServerError, "Error decoding parameters for user creation")
		return
	}
	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	var userInfo database.CreateUserParams = database.CreateUserParams{
		Email: payload.Email,
		Password: hashedPassword,
	}


	user, err := cfg.dbQueries.CreateUser(req.Context(), userInfo)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	newUser := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}


	jsonResponse(w, http.StatusCreated, newUser)
}