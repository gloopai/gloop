package auth

import "github.com/gloopai/gloop/modules/db"

type AuthOptions struct {
	Db         *db.DbService
	JWTOptions JWTOptions `json:"jwt_options"` // JWT 选项
}
