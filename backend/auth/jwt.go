package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenGenerator interface {
	GenerateToken(userID int64) (string, error)
}

type DefaultTokenGenerator struct{}

func (d DefaultTokenGenerator) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user.id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

var jwtSecret = []byte("your_secret_key")

func ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
}
