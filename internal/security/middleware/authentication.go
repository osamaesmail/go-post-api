package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/osamaesmail/go-post-api/internal/config"
	"github.com/osamaesmail/go-post-api/internal/constant"
	"github.com/osamaesmail/go-post-api/internal/web"
)

func JWTVerifier(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get(constant.API_KEY_HEADER)
		if tokenHeader == "" {
			web.MarshalError(w, http.StatusUnauthorized, constant.ErrUnauthorized)
			return
		}

		tokenParse, err := jwt.Parse(tokenHeader, func(jwtToken *jwt.Token) (interface{}, error) {
			if jwtToken.Method != jwt.SigningMethodHS256 {
				return nil, constant.ErrUnauthorized
			}
			return []byte(config.Cfg().JwtSecretKey), nil
		})

		if err != nil || !tokenParse.Valid {
			web.MarshalError(w, http.StatusUnauthorized, constant.ErrUnauthorized)
			return
		}

		claims := tokenParse.Claims.(jwt.MapClaims)
		claimsID, err := strconv.ParseInt(fmt.Sprint(claims["id"]), 10, 64)
		if err != nil {
			web.MarshalError(w, http.StatusUnauthorized, constant.ErrUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsIDKey, claimsID)))
	})
}
