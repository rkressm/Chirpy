package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rkressm/Chirpy/internal/auth"
	"github.com/rkressm/Chirpy/internal/database"
)

type parameters struct {
	Body    string    `json:"body"`
	User_id uuid.UUID `json:"user_id"`
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	param, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, "error parsing param")
		return
	}
	chirp, err := cfg.db.GetChirp(r.Context(), param)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "token does not exist")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "error validating user")
		return
	}
	if chirp.UserID != userID {
		respondWithError(w, 403, "not allowed to delete")
		return
	}
	err = cfg.db.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, 500, "error deleting chirp")
		return
	}
	w.WriteHeader(204)
}

func (cfg *apiConfig) handlerRetrieveChirps(w http.ResponseWriter, r *http.Request) {
	var chirps []database.Chirp
	var err error
	var author_id uuid.UUID
	s := r.URL.Query().Get("author_id")
	order := r.URL.Query().Get("sort")
	fmt.Println("order:", order)
	if s == "" {
		chirps, err = cfg.db.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, 500, "error fetching chirps")
			return
		}
	} else {
		author_id, err = uuid.Parse(s)
		if err != nil {
			respondWithError(w, 500, "error converting id")
			return
		}
		chirps, err = cfg.db.GetChirpByAuthorId(r.Context(), author_id)
		if err != nil {
			respondWithError(w, 404, "chirp not found")
			return
		}
	}
	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	resp := []response{}
	for _, chirp := range chirps {
		resp = append(resp, response{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}
	if order == "desc" {
		sort.Slice(resp, func(i, j int) bool {
			return resp[i].CreatedAt.After(resp[j].CreatedAt)
		})
	} else {
		sort.Slice(resp, func(i, j int) bool {
			return resp[i].CreatedAt.Before(resp[j].CreatedAt)
		})
	}
	respondWithJSON(w, 200, resp)
}

func (cfg *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	param, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, "error parsing param")
	}

	chirp, err := cfg.db.GetChirp(r.Context(), param)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}
	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	resp := response{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, 200, resp)
}

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Invalid token")
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "Invalid token")
		return
	}
	err = decoder.Decode(&params)
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
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error while creating chirp")
		return
	}
	resp := response{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, 201, resp)
}
