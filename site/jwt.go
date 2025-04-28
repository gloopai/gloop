package site

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type AuthJwtClaims struct {
	UserId   int64  `json:"user_id"`
	UserName string `json:"username"`
	jwt.StandardClaims
}

type JWTOptions struct {
	SecretKey     string        `json:"secret_key"`     // Secret key for signing JWT tokens
	TokenDuration time.Duration `json:"token_duration"` // Duration for which the token is valid
}

func NewJWTManager(opt JWTOptions) *JWTManager {
	return &JWTManager{
		secretKey:     opt.SecretKey,
		tokenDuration: opt.TokenDuration,
	}
}

func (j *JWTManager) GenerateToken(auth RequestAuth) (string, error) {
	claims := AuthJwtClaims{
		UserId:   auth.UserId,
		UserName: auth.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(j.tokenDuration).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTManager) VerifyToken(tokenString string) (RequestAuth, error) {
	claims := &AuthJwtClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return RequestAuth{}, nil
	})

	if err != nil {
		return RequestAuth{}, err
	}

	if !token.Valid {
		return RequestAuth{}, jwt.ErrSignatureInvalid
	}

	return RequestAuth{
		UserId:   claims.UserId,
		Username: claims.UserName,
	}, nil
}
