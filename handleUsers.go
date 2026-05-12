package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type email struct {
	Email string `json:"email"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	email := email{}
	type response struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	err := decoder.Decode(&email)
	if err != nil {
		log.Printf("Error decoding parameter: %s", err)
		respondWithError(w, 400, "Error decoding parameter")
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), email.Email)
	if err != nil {
		log.Printf("error creating user: %s", err)
		return
	}
	resp := response{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, 201, resp)
}
