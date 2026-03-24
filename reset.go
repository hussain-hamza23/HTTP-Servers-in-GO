package main

import (
	"httpserver/internal/auth"
	"log"
	"net/http"
	"time"
)

const roleDev string = "dev"

// func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request){
// 	cfg.fileserverHits.Store(0)
// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 	if _, err := w.Write([]byte("Hits reset to 0")); err != nil {
// 		httpError := &HTTPError{
// 			StatusCode: http.StatusInternalServerError,
// 			Message: ErrResponse,
// 			Err: err,
// 		}
// 		httpError.errorResponse(w)
// 		return
// 	}
// }

func (cfg *apiConfig) resetUsersHandler(w http.ResponseWriter, r *http.Request){
	if cfg.role != roleDev{
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
		log.Printf("error writing response: %v", err)
		return
	}
}

func (cfg *apiConfig) refreshTokensHandler(w http.ResponseWriter, req *http.Request){
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: "Missing or invalid refresh token",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}

	user, err := cfg.dbQueries.GetUserFromValidRefreshToken(req.Context(), refreshToken)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: (ErrUserNotFound + ": invalid refresh token"),
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}
	newToken, err := auth.MakeJWT(user.ID, string(cfg.tokenSecret), time.Hour)
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
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: "Missing or invalid refresh token",
			Err: err,
		}
		httpError.errorResponse(w)
		return
	}
	var e error = cfg.dbQueries.RevokeRefreshToken(req.Context(), refreshToken) 
	if e != nil {
		httpError := &HTTPError{
			StatusCode: http.StatusInternalServerError,
			Message: "Error revoking token",
			Err: e,
		}
		httpError.errorResponse(w)
		return
	}

	jsonResponse(w, http.StatusNoContent, nil)
}