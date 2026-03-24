package main

import (
	"httpserver/internal/auth"
	"httpserver/internal/database"
	"net/http"
	"sort"
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

	payload, e := decodeRequest[requestPayload](w, req)
	if e != nil{
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

func getAllChirpsHandler(w http.ResponseWriter, req *http.Request, cfg *apiConfig) ([]chirpPayload, *HTTPError){
	chirps, err := cfg.dbQueries.GetAllChirps(req.Context())
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error fetching chirps",
			Err: err,
		}
		httpError.errorResponse(w)
		return nil, httpError
	}

	var response []chirpPayload = make([]chirpPayload,0, len(chirps))
	for _, chirp := range chirps{
		response = append(response, newChirpPayload(chirp))
	}
	return response, nil
}

func (cfg *apiConfig)getChirpsHandler(w http.ResponseWriter, req *http.Request){
	var authorQuery string = req.URL.Query().Get("author_id")
	var sortQuery string = req.URL.Query().Get("sort")
	if sortQuery == ""{
		sortQuery = "asc"
	}

	if authorQuery == ""{
		chirps, err := getAllChirpsHandler(w, req, cfg)
		if err != nil {
			return
		}
		if sortQuery == "desc"{
			sort.Slice(chirps, func(i, j int) bool{
				return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
			})
		} else{
			sort.Slice(chirps, func(i, j int) bool{
				return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
			})
		}
		jsonResponse(w, http.StatusOK, chirps)
		return
	}

	authorID, err := uuid.Parse(authorQuery)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: "Invalid UUID",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	chirps, err := cfg.dbQueries.GetChirpsByAuthorID(req.Context(), authorID)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusNotFound,
			Message: "Chirp not found",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	var response []chirpPayload = make([]chirpPayload, 0, len(chirps))
	for _, chirp := range chirps {
		response = append(response, newChirpPayload(chirp))
	}
	if sortQuery == "desc"{
		sort.Slice(response, func(i, j int) bool{
			return response[i].CreatedAt.After(response[j].CreatedAt)
		})
	} else{
		sort.Slice(response, func(i, j int) bool{
			return response[i].CreatedAt.Before(response[j].CreatedAt)
		})
	}


	jsonResponse(w, http.StatusOK, response)
}

func (cfg *apiConfig)deleteChirpHandler(w http.ResponseWriter, req *http.Request){
	id, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: "Invalid UUID",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	accessToken, err := auth.GetBearerToken(req.Header) 
	if err != nil{
		httpError := &HTTPError{
			StatusCode: http.StatusUnauthorized,
			Message: "Could not retrieve access token",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, string(cfg.tokenSecret))
	if err != nil{
		httpError := &HTTPError{
			StatusCode: http.StatusUnauthorized,
			Message: ErrInvalidToken,
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	deleteChirpParams := database.DeleteChirpFromIDParams{
		ID: id,
		UserID: userID,
	}

	result, err := cfg.dbQueries.DeleteChirpFromID(req.Context(), deleteChirpParams)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error deleting chirp",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	if rowsDeleted, err := result.RowsAffected(); err != nil || rowsDeleted == 0{
		httpError := &HTTPError{
			StatusCode: http.StatusForbidden,
			Message: "User does not have permission to delete this chirp",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}