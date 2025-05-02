package auth

import (
	"time"

	"github.com/gloopai/gloop/modules"
	"github.com/golang-jwt/jwt/v4"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type AuthJwtClaims struct {
	UserId   int64  `json:"user_id"`
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

type JWTOptions struct {
	SecretKey     string `json:"secret_key"`     // Secret key for signing JWT tokens
	Authorization string `json:"authorization"`  // Authorization header name
	TokenDuration int    `json:"token_duration"` // Duration for which the token is valid
}

func NewJWTManager(opt JWTOptions) *JWTManager {
	return &JWTManager{
		secretKey:     opt.SecretKey,
		tokenDuration: time.Duration(time.Hour * time.Duration(opt.TokenDuration)),
	}
}

func (j *JWTManager) GenerateToken(auth modules.RequestAuth) (string, error) {
	claims := AuthJwtClaims{
		UserId:   auth.UserId,
		UserName: auth.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTManager) VerifyToken(tokenString string) (modules.RequestAuth, error) {
	claims := &AuthJwtClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return modules.RequestAuth{}, nil
	})

	if err != nil {
		return modules.RequestAuth{}, err
	}

	if !token.Valid {
		return modules.RequestAuth{}, jwt.ErrSignatureInvalid
	}

	return modules.RequestAuth{
		UserId:   claims.UserId,
		Username: claims.UserName,
	}, nil
}
