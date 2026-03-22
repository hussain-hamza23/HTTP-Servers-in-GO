package main

import (
	"encoding/json"
	"httpserver/internal/database"
	"net/http"
	"time"

	"github.com/google/uuid"
)
type chirpPayload struct{
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body string `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}


func (cfg *apiConfig)createChirpHandler(w http.ResponseWriter, req *http.Request){
	type requestPayload struct{
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}



	var payload requestPayload = requestPayload{}
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := decoder.Decode(&payload); err != nil{
		errorResponse(w, http.StatusInternalServerError, "Error decoding parameters for chirp creation")
		return
	}

	body, err := validateLengthHandler(w, payload.Body)
	if err != nil{
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(req.Context(),database.CreateChirpParams{
		Body: body,
		UserID: payload.UserID,
	})
	if err != nil{
		errorResponse(w, http.StatusInternalServerError, "Error creating chirp")
		return
	}

	newChirp := chirpPayload{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}

	jsonResponse(w, http.StatusCreated, newChirp)
}

func (cfg *apiConfig)getAllChirpsHandler(w http.ResponseWriter, req *http.Request){
	chirps, err := cfg.dbQueries.GetAllChirps(req.Context())
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Error fetching chirps")
		return
	}

	var response []chirpPayload = make([]chirpPayload,0, len(chirps))
	for _, chirp := range chirps{
		response = append(response, chirpPayload{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}
	jsonResponse(w, http.StatusOK, response)
}

func (cfg *apiConfig)getSingleChirpHandler(w http.ResponseWriter, req *http.Request){
	id, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid UUID")
		return
	}

	chirp, err := cfg.dbQueries.GetSingleChirp(req.Context(), id)
	if err != nil {
		errorResponse(w, http.StatusNotFound, "Chirp not found")
		return
	}

	response := chirpPayload{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}

	jsonResponse(w, http.StatusOK, response)
}