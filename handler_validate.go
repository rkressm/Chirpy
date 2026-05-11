package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

type parameters struct {
	Body string `json:"body"`
}

func profanitySanitizer(params *parameters) {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	splittedMsg := strings.Fields(params.Body)
	for i, word := range splittedMsg {
		if slices.Contains(badWords, strings.ToLower(word)) == true {
			splittedMsg[i] = "****"
		}
	}
	params.Body = strings.Join(splittedMsg, " ")
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	type response struct {
		Value string `json:"cleaned_body"`
	}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameter: %s", err)
		respondWithError(w, 500, "Error decoding parameter")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	profanitySanitizer(&params)
	resp := response{
		Value: params.Body,
	}
	respondWithJSON(w, 200, resp)
}
