package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func errorResponse(w http.ResponseWriter, statusCode int, message string){
	type errorResponse struct{
		Error string `json:"error"`
	}

	response := errorResponse{
		Error: message,
	}
	jsonResponse(w, statusCode, response)
}

func jsonResponse(w http.ResponseWriter, statusCode int, data interface{}){
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	dat, err := json.Marshal(data)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error encoding response: %s", err))
		return
	}
	if _, err := w.Write(dat); err != nil {
		log.Printf("Error writing JSON response: %s\n", err)
	}
}

func validateLengthHandler(w http.ResponseWriter, req *http.Request){
	type requestPayload struct {
		Body string `json:"body"`
	}

	// Decode the JSON payload from the request body
	var decoder *json.Decoder = json.NewDecoder(req.Body)
	defer req.Body.Close()
	// Initialize an empty payload struct to hold the decoded data
	var payload requestPayload = requestPayload{}

	if err := decoder.Decode(&payload); err != nil{
		errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error decoding parameters: %s", err))
		return
	}

	const maxLength int = 140

	if len(payload.Body) > maxLength{
		errorResponse(w, http.StatusBadRequest, fmt.Sprintf("Body exceeds %d characters", maxLength))
		return
	}

	var cleanedBody string = profanityChecker(payload.Body)
	jsonResponse(w, http.StatusOK, map[string]string{"cleaned_body": cleanedBody})
}


func profanityChecker(body string) string{
	var profanityList map[string]struct{} = map[string]struct{}{
		"kerfuffle": {},
		"sharbert": {},
		"fornax": {},
	}
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