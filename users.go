package main

import (
	"httpserver/internal/auth"
	"httpserver/internal/database"
	"net/http"
	"time"

	"github.com/google/uuid"
	//_ "github.com/lib/pq"
)

type User struct{
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email string `json:"email"`
	Password string `json:"password"`
	IsChirpyRed bool `json:"is_chirpy_red"`
	Token string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, req *http.Request){
	type requestPayload struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}

	payload, e := decodeRequest[requestPayload](w, req)
	if e != nil{
		return
	}
	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error hashing password",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	var userInfo database.CreateUserParams = database.CreateUserParams{
		Email: payload.Email,
		Password: hashedPassword,
	}


	user, err := cfg.dbQueries.CreateUser(req.Context(), userInfo)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error creating user",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	newUser := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
	jsonResponse(w, http.StatusCreated, newUser)
}


func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, req *http.Request){

}


func (cfg *apiConfig) updateUserPasswordHandler(w http.ResponseWriter, req *http.Request){
	type requestPayload struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}
	
	payload, e := decodeRequest[requestPayload](w, req)
	if e != nil{
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

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil{
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error hashing password",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	updatedEmailAndPassword := database.UpdateEmailandPasswordParams{
		Email: payload.Email,
		Password: hashedPassword,
		ID: userID,
	}

	updatedUser, err := cfg.dbQueries.UpdateEmailandPassword(req.Context(), updatedEmailAndPassword)
	if err != nil{
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error updating user information",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	updatedUserResponse := User{
		ID: updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email: updatedEmailAndPassword.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	}
	jsonResponse(w, http.StatusOK, updatedUserResponse)
}
