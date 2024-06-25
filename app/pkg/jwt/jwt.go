package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/seakee/go-api/app"
	"github.com/seakee/go-api/app/model/auth"
)

type ServerClaims struct {
	ID      uint   `json:"id"`
	AppName string `json:"app_name"`
	AppID   string `json:"app_id"`
	jwt.RegisteredClaims
}

// GenerateAppToken 获取app应用的token
// expireTime 过期时间，单位秒
func GenerateAppToken(App *auth.App, expireTime time.Duration) (token string, err error) {
	expTime := time.Now().Add(expireTime * time.Second)
	claims := ServerClaims{
		ID:      App.ID,
		AppName: App.AppName,
		AppID:   App.AppID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-api",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := []byte(app.GetConfig().System.JwtSecret)

	return tokenClaims.SignedString(jwtSecret)
}

func ParseAppAuth(token string) (*ServerClaims, error) {
	jwtSecret := []byte(app.GetConfig().System.JwtSecret)

	tokenClaims, err := jwt.ParseWithClaims(token, &ServerClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*ServerClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}
