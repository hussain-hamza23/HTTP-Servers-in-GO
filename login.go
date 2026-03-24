package main

import (
	"context"
	"encoding/json"
	"httpserver/internal/auth"
	"httpserver/internal/database"
	"net/http"
	"time"
)

type requestPayload struct{
	Email string `json:"email"`
	Password string `json:"password"`
}



func (cfg *apiConfig) loginHandler(w http.ResponseWriter, req *http.Request){

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

	user, err := cfg.getUserInfo(req.Context(), payload)
	if err != nil{
		err.errorResponse(w)
		return
	}

	if err := cfg.generateTokens(&user, req.Context()); err != nil{
		err.errorResponse(w)
		return
	}


	jsonResponse(w, http.StatusOK, user)
}

func (cfg *apiConfig) getUserInfo(ctx context.Context, payload requestPayload) (User, *HTTPError){
	user, err := cfg.dbQueries.GetUserByEmail(ctx, payload.Email)
	if err != nil{
		return User{}, &HTTPError{
			StatusCode: http.StatusUnauthorized,
			Message: (ErrInternalServer + ": user not found"),
			Err: err,
		}
	}
	match, err := auth.CheckPasswordHash(payload.Password, user.Password)
	if err != nil{
		return User{}, &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: (ErrInternalServer + ": error checking password"),
			Err: err,
		}
	}
	if !match {
		return User{}, &HTTPError{
			StatusCode: http.StatusUnauthorized,
			Message: ErrInvalidCredentials,
		}
	}
	var userResponse User = User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	
	return userResponse, nil
}

func (cfg *apiConfig) generateTokens(user *User, ctx context.Context) *HTTPError{
	if err := cfg.generateJWT(user); err != nil {
		return err
	}
	if err := cfg.generateRefreshToken(user, ctx); err != nil {
		return err
	}
	return nil
}

func (cfg *apiConfig) generateJWT(user *User) *HTTPError {
	token, err := auth.MakeJWT(user.ID, string(cfg.tokenSecret), time.Hour)
	if err != nil {
		return &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: (ErrInternalServer + ": error generating token"),
			Err: err,
		}
	}
	user.Token = token
	return nil
}

func (cfg *apiConfig) generateRefreshToken(user *User, ctx context.Context) *HTTPError{
	var refreshTokenParams database.CreateRefreshTokenParams = database.CreateRefreshTokenParams{
	UserID: user.ID,
	Token: auth.MakeRefreshToken(),
	ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	}
	refreshToken, err := cfg.dbQueries.CreateRefreshToken(ctx, refreshTokenParams)
	if err != nil {
		return &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: (ErrInternalServer + ": error creating refresh token"),
			Err: err,
		}
	}
	user.RefreshToken = refreshToken.Token
	return nil
}