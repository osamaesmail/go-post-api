package token

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/osamaesmail/go-post-api/internal/config"
)

type Generator interface {
	GenerateClaims() jwt.MapClaims
}

func GenerateToken(g Generator) (string, error) {
	claims := g.GenerateClaims()
	claims["exp"] = time.Now().Add(config.Cfg().JwtTTL).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Cfg().JwtSecretKey))
}
