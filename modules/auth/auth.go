package auth

import (
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules"
	"github.com/gloopai/gloop/modules/db"
)

type RequestAuth struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
}

type Auth struct {
	modules.Base
	Config     AuthOptions // 认证配置
	db         *db.Db
	JWTManager *JWTManager // JWT 管理器
}

func NewAuth(opt AuthOptions) *Auth {
	return &Auth{
		db:     opt.Db,
		Config: opt,
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

	if a.Config.JWTOptions.Authorization == "" {
		a.Config.JWTOptions.Authorization = "Authorization"
	}

	a.JWTManager = NewJWTManager(a.Config.JWTOptions)
}

func (a *Auth) Start() error {
	return nil
}

/* 获取用户表名 */
func (a *Auth) TableName() string {
	return new(User).TableName()
}

/* 用户注册 */
func (a *Auth) Register(user *User) error {
	return RegisterUser(a.db.Db, user.Username, user.Password, user.Email)
}

/* 用户登录 */
func (a *Auth) Login(user *User) error {
	loggedInUser, err := LoginUser(a.db.Db, user.Username, user.Password)
	if err != nil {
		return err
	}

	// Populate the user details
	*user = *loggedInUser
	return nil
}
