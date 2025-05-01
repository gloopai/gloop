package auth

import (
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules"
	"github.com/gloopai/gloop/modules/db"
)

type Auth struct {
	modules.Base
	db *db.Db
}

func NewAuth(opt AuthOptions) *Auth {
	return &Auth{
		db: opt.Db,
	}
}
func (a *Auth) Name() string {
	return "auth"
}
func (a *Auth) Init() {
	err := EnsureAuthTableExists(a.db.Db)
	if err != nil {
		lib.Log.Error("Failed to ensure auth table exists:", err)
		return
	}
}

func (a *Auth) Start() error {
	return nil
}
