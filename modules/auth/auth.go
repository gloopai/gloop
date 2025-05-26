package auth

import (
	"fmt"

	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules"
	"github.com/gloopai/gloop/modules/db"
)

type Auth struct {
	modules.Base
	Config     AuthOptions // 认证配置
	db         *db.DbService
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

/* 获取 header 中的 Authorization key*/
func (a *Auth) Authorization() string {
	return a.Config.JWTOptions.Authorization
}

/* 用户注册 */
func (a *Auth) Register(req *modules.RequestPayload) modules.ResponsePayload {
	type queryObj struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	var query queryObj
	err := req.Unmarshal(&query)
	if err != nil {
		return modules.Response.Error(err.Error())
	}
	err = RegisterUser(a.db.Db, query.Username, query.Password, query.Email)
	if err != nil {
		return modules.Response.Error(err.Error())
	}

	return modules.Response.SuccessNone()
}

/* 用户登录 */
func (a *Auth) Login(req *modules.RequestPayload) modules.ResponsePayload {
	type queryObject struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var query queryObject
	err := req.Unmarshal(&query)
	if err != nil {
		return modules.Response.Error(err.Error())
	}

	loggedInUser, err := LoginUser(a.db.Db, query.Username, query.Password)
	if err != nil {
		return modules.Response.Error(err.Error())
	}

	token, err := a.JWTManager.GenerateToken(modules.RequestAuth{
		UserId:   loggedInUser.Id,
		Username: loggedInUser.Username,
	})
	// Populate the user details
	if err != nil {
		return modules.Response.Error(err.Error())
	}

	// loggedInUser.Token = token
	resmap := make(map[string]interface{})
	resmap["token"] = token

	return modules.Response.Success(resmap)
}

/* 获取用户信息 */
func (a *Auth) ParseToken(req *modules.RequestPayload) modules.ResponsePayload {
	type queryObject struct {
		Token string `json:"token"`
	}
	var query queryObject
	err := req.Unmarshal(&query)
	if err != nil {
		return modules.Response.Error(err.Error())
	}
	auth, err := a.JWTManager.VerifyToken(query.Token)
	if err != nil {
		return modules.Response.Error(fmt.Sprintf("JWTERROR:%s", err.Error()))
	}

	return modules.Response.Success(auth)
}
