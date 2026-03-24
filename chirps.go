package main

import (
	"encoding/json"
	"httpserver/internal/auth"
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

func newChirpPayload(chirp database.Chirp) chirpPayload{
	return chirpPayload{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}
}


func (cfg *apiConfig)createChirpHandler(w http.ResponseWriter, req *http.Request){
	type requestPayload struct{
		Body string `json:"body"`
	}

	var payload requestPayload = requestPayload{}
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := decoder.Decode(&payload); err != nil{
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: ErrDecodingRequest,
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusUnauthorized,
			Message: ErrInvalidToken,
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	userID, err := auth.ValidateJWT(token, string(cfg.tokenSecret))
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusUnauthorized,
			Message: ErrInvalidToken,
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	body, err := validateLengthHandler(w, payload.Body)
	if err != nil{
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(req.Context(),database.CreateChirpParams{
		Body: body,
		UserID: userID,
	})
	if err != nil{
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error creating chirp",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	newChirp := newChirpPayload(chirp)

	jsonResponse(w, http.StatusCreated, newChirp)
}

func (cfg *apiConfig)getAllChirpsHandler(w http.ResponseWriter, req *http.Request){
	chirps, err := cfg.dbQueries.GetAllChirps(req.Context())
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error fetching chirps",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	var response []chirpPayload = make([]chirpPayload,0, len(chirps))
	for _, chirp := range chirps{
		response = append(response, newChirpPayload(chirp))
	}
	jsonResponse(w, http.StatusOK, response)
}

func (cfg *apiConfig)getSingleChirpHandler(w http.ResponseWriter, req *http.Request){
	id, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: "Invalid UUID",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	chirp, err := cfg.dbQueries.GetSingleChirp(req.Context(), id)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusNotFound,
			Message: "Chirp not found",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	response := newChirpPayload(chirp)

	jsonResponse(w, http.StatusOK, response)
}