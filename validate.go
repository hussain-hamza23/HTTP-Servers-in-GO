package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type HTTPError struct{
	StatusCode int
	Message string
	Err error
}
func (err *HTTPError) Error() string{
	return err.Message
}

const (
	ErrInvalidCredentials = "incorrect email or password"
	ErrUserNotFound = "user not found"
	ErrInvalidToken = "invalid token"
	ErrInternalServer = "internal server error"
	ErrResponse = "error writing response"

	ErrEncodingResponse = "error encoding response"
	ErrDecodingRequest = "error decoding request"
)

var profanityList map[string]struct{} = map[string]struct{}{
	"kerfuffle": {},
	"sharbert": {},
	"fornax": {},
}

func (httpError *HTTPError) errorResponse(w http.ResponseWriter){
	type errorResponse struct{
		ErrorMessage string `json:"error"`
		Error string `json:"error_details,omitempty"`
	}
	var errDetails string
	if httpError.Err != nil {
		errDetails = httpError.Err.Error()
	}

	response := errorResponse{
		ErrorMessage: httpError.Error(),
		Error: errDetails,
	}
	jsonResponse(w, httpError.StatusCode, response)
}

func jsonResponse(w http.ResponseWriter, statusCode int, data any) {
    dat, err := json.Marshal(data)
    if err != nil {
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"error":"` + ErrEncodingResponse + `"}`))
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(statusCode)
    if _, err := w.Write(dat); err != nil {
		//w.Write failure is unrecoverable
        log.Printf(ErrResponse + ": %v", err)
    }
}

// func validateLengthHandler(w http.ResponseWriter, req *http.Request) (string, error) {
// 	type requestPayload struct {
// 		Body string `json:"body"`
// 	}

// 	// Decode the JSON payload from the request body
// 	var decoder *json.Decoder = json.NewDecoder(req.Body)
// 	defer req.Body.Close()
// 	// Initialize an empty payload struct to hold the decoded data
// 	var payload requestPayload = requestPayload{}

// 	if err := decoder.Decode(&payload); err != nil{
// 		errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error decoding parameters for chirp creation: %s", err))
// 		return "", err
// 	}

// 	const maxLength int = 140

// 	if len(payload.Body) > maxLength{
// 		errorResponse(w, http.StatusBadRequest, fmt.Sprintf("Body exceeds %d characters", maxLength))
// 		return "", errors.New("Body exceeds maximum length")
// 	}

// 	var cleanedBody string = profanityChecker(payload.Body)
// 	// jsonResponse(w, http.StatusOK, map[string]string{"cleaned_body": cleanedBody})
// 	return cleanedBody, nil
// }


func validateLengthHandler(w http.ResponseWriter, chirp string) (string, error) {
	const maxLength int = 140

	if len(chirp) > maxLength{
		httpError := &HTTPError{
			StatusCode: http.StatusBadRequest,
			Message: fmt.Sprintf("Body exceeds %d characters", maxLength),
		}
		httpError.errorResponse(w)
		return "", errors.New("Body exceeds maximum length")
	}
	var cleanedBody string = profanityChecker(chirp)
	return cleanedBody, nil
}



func profanityChecker(body string) string{

	const replacement string = "****"
	var splitText []string = strings.Split(body, " ")

	for idx, word := range splitText{
		check := strings.ToLower(word)
		if _, exists := profanityList[check]; exists{
			splitText[idx] = replacement
		}
	}
	return strings.Join(splitText, " ")
}