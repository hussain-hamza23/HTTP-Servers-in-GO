package main

import (
	"fmt"
	"httpserver/internal/auth"
	"net/http"
	"time"
)

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request){
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte("Hits reset to 0")); err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: ErrResponse,
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}
}

func (cfg *apiConfig) resetUsersHandler(w http.ResponseWriter, r *http.Request){
	if cfg.role != "dev"{
		httpError := &HTTPError{
			StatusCode: http.StatusForbidden,
			Message: "Forbidden: insufficient permissions",
		}
		httpError.errorResponse(w)
		return
	}
	if err := cfg.dbQueries.DeleteUsers(r.Context()); err != nil{
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error resetting users",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte("Users reset successfully")); err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: ErrResponse,
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}
}

func (cfg *apiConfig) refreshTokensHandler(w http.ResponseWriter, req *http.Request){
	var refreshToken string = req.Header.Get("Authorization")
	if refreshToken == "" {
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: "Missing refresh token",
		}
		httpError.errorResponse(w)
		return
	}

	i, err := fmt.Sscanf(refreshToken, "Bearer %s", &refreshToken)
	if err != nil || i != 1{
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: "Invalid refresh token format",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	isValid, err := cfg.dbQueries.CheckTokenStatus(req.Context(), refreshToken)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error checking token status",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	if !isValid {
		httpError := &HTTPError{
			StatusCode: http.StatusUnauthorized,
			Message: ErrInvalidToken,
		}
		httpError.errorResponse(w)
		return
	}

	user, err := cfg.dbQueries.GetUserFromRefreshToken(req.Context(), refreshToken)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: (ErrUserNotFound + ": invalid refresh token"),
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}
	newToken, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Hour)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: (ErrInternalServer + ": error generating token"),
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}
	refresh := struct{

		Token string `json:"token"`
	}{
		Token: newToken,
	}

	jsonResponse(w, http.StatusOK, refresh)
}


func (cfg *apiConfig) revokeTokenHandler(w http.ResponseWriter, req *http.Request){
	var refreshToken string = req.Header.Get("Authorization")
	if refreshToken == "" {
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: "Missing refresh token",
		}
		httpError.errorResponse(w)
		return
	}

	i, err := fmt.Sscanf(refreshToken, "Bearer %s", &refreshToken)
	if err != nil || i != 1{
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: "Invalid refresh token format",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	var revoke error = cfg.dbQueries.RevokeRefreshToken(req.Context(), refreshToken) 
	if revoke != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error revoking token",
			Err: revoke,
		}
		httpError.errorResponse(w)
		return
	}

	jsonResponse(w, http.StatusNoContent, nil)
}