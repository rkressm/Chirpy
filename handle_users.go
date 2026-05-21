package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rkressm/Chirpy/internal/auth"
	"github.com/rkressm/Chirpy/internal/database"
)

type users struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (cfg *apiConfig) handlerUpdateToRed(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, 401, "apiKey not found")
		return
	}
	if apiKey != cfg.polkaKey {
		w.WriteHeader(401)
		return
	}
	type param struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	params := param{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Error decoding parameter")
		return
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	err = cfg.db.SetChirpyRedByID(r.Context(), params.Data.UserID)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(204)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := users{}
	type response struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		IsRed     bool      `json:"is_chirpy_red"`
	}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameter: %s", err)
		respondWithError(w, 400, "Error decoding parameter")
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "token does not exist")
		return
	}
	id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "error validating user")
		return
	}
	params.Password, err = auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 401, "error hashing pwd")
		return
	}

	updatedUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: params.Password,
		ID:             id,
	})
	if err != nil {
		respondWithError(w, 401, "error updating user")
		return
	}
	resp := response{
		ID:        updatedUser.ID.String(),
		UpdatedAt: updatedUser.UpdatedAt,
		CreatedAt: updatedUser.CreatedAt,
		Email:     updatedUser.Email,
		IsRed:     updatedUser.IsChirpyRed.Bool,
	}
	respondWithJSON(w, 200, resp)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "token does not exist")
		return
	}
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 401, "no returned user")
		return
	}
	activeToken, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, 401, "error generating token")
		return
	}
	resp := response{
		Token: activeToken,
	}
	respondWithJSON(w, 200, resp)
}
func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "token does not exist")
		return
	}
	err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 500, "error revoking token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := users{}
	type response struct {
		ID           string    `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsRed        bool      `json:"is_chirpy_red"`
	}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameter: %s", err)
		respondWithError(w, 400, "Error decoding parameter")
		return
	}
	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("error getting user: %s", err)
		return
	}
	isValid, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "error checking pwd hash")
		return
	}
	if !isValid {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	activeToken, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	refreshToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     auth.MakeRefresherToken(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		respondWithError(w, 500, "error creating a refresh token")
	}
	resp := response{
		ID:           user.ID.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        activeToken,
		RefreshToken: refreshToken.Token,
		IsRed:        user.IsChirpyRed.Bool,
	}
	respondWithJSON(w, 200, resp)
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := users{}
	type response struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		IsRed     bool      `json:"is_chirpy_red"`
	}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameter: %s", err)
		respondWithError(w, 400, "Error decoding parameter")
		return
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("error creating user: %s", err)
		return
	}
	resp := response{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		IsRed:     user.IsChirpyRed.Bool,
	}
	respondWithJSON(w, 201, resp)
}
