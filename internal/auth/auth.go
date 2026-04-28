package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const tokenExp = time.Minute * 15
const secretKey = "supersecretkey"

var secretKeyBytes = []byte(secretKey)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// BuildJWTString generates and signs a JWT string for the given user ID using HS256 and returns it or an error.
func BuildJWTString(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID.String(),
	})

	tokenString, err := token.SignedString(secretKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

// GetUserID extracts the user ID from the JWT string and returns it or an error.
func GetUserID(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return secretKeyBytes, nil
		})

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user id: %w", err)
	}

	return userID, nil
}
