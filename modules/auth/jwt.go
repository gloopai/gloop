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
	if opt.SecretKey == "" {
		opt.SecretKey = "RxyiJcD8O19/GE9GL/V2sn0b/MOSWTWoygN77e7RNSI="
	}

	if opt.TokenDuration == 0 {
		opt.TokenDuration = 24 * 365 // Default token duration is 24 hours
	}

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
		// 验证签名方法是否为 HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		// 返回用于验证签名的密钥
		return []byte(j.secretKey), nil
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
