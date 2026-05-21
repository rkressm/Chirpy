package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rkressm/Chirpy/internal/auth"
)

func TestJWTThing(t *testing.T) {
	t.Run("valid operation", func(t *testing.T) {
		userID := uuid.New()
		secret := "coucou"

		tokenString, err := auth.MakeJWT(userID, secret, 5*time.Second)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		gotID, err := auth.ValidateJWT(tokenString, secret)

		if err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}

		if gotID != userID {
			t.Errorf("expected ID %v, got %v", userID, gotID)
		}
	})
	t.Run("invalid time", func(t *testing.T) {
		userID := uuid.New()
		secret := "coucou"

		tokenString, err := auth.MakeJWT(userID, secret, -5*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = auth.ValidateJWT(tokenString, secret)
		if err == nil {
			t.Fatalf("expected validation error: %v", err)
		}
	})
	t.Run("invalid secret", func(t *testing.T) {
		userID := uuid.New()
		secret := "coucou"
		tokenString, err := auth.MakeJWT(userID, secret, 5*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = auth.ValidateJWT(tokenString, "wrong-secret")
		if err == nil {
			t.Fatalf("expected validation error: %v", err)
		}
	})
}
