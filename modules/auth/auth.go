package auth

import (
	"context"
	"fmt"

	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules"
)

type Auth struct {
	modules.Base
	Config     AuthOptions // 认证配置
	JWTManager *JWTManager // JWT 管理器
}

func NewAuth(opt AuthOptions) *Auth {
	return &Auth{
		Config: opt,
	}
}
func (a *Auth) Name() string {
	return "auth"
}
func (a *Auth) Init() {
	err := EnsureAuthTableExists(a.Env.Mysql.Conn)
	if err != nil {
		lib.Log.Error("Failed to ensure auth table exists:", err)
		return
	}

	if a.Config.JWTOptions.Authorization == "" {
		a.Config.JWTOptions.Authorization = "Authorization"
	}

	a.JWTManager = NewJWTManager(a.Config.JWTOptions)

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
func (a *Auth) Register(ctx context.Context, req *modules.RequestPayload) modules.ResponsePayload {
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
	err = RegisterUser(a.Env.Mysql.Conn, query.Username, query.Password, query.Email)
	if err != nil {
		return modules.Response.Error(err.Error())
	}

	return modules.Response.SuccessNone()
}

/* 用户登录 */
func (a *Auth) Login(ctx context.Context, req *modules.RequestPayload) modules.ResponsePayload {
	type queryObject struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var query queryObject
	err := req.Unmarshal(&query)
	if err != nil {
		return modules.Response.Error(err.Error())
	}

	loggedInUser, err := LoginUser(a.Env.Mysql.Conn, query.Username, query.Password)
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

// 通过 Telegram 登录
func (a *Auth) LoginByTelegram(ctx context.Context, req *modules.RequestPayload) modules.ResponsePayload {
	type queryObject struct {
		InitData string `json:"init_data"`
	}
	var query queryObject
	err := req.Unmarshal(&query)
	if err != nil {
		return modules.Response.Error(err.Error())
	}
	telegramUser := &TelegramUser{}
	_, err = telegramUser.Login(a.Env.Mysql.Conn, query.InitData, a.Config.TelegramBotToken)
	if err != nil {
		return modules.Response.Error(fmt.Sprintf("Telegram parse error: %s", err.Error()))
	}

	token, err := a.JWTManager.GenerateToken(modules.RequestAuth{
		UserId:   telegramUser.UserId,
		Username: "", // TelegramUser struct does not have a Username field, you might want to add it or handle differently
	})

	// telegramUser, err := LoginUserByTelegram(a.db.Db, query.InitData)
	return modules.Response.Success(map[string]interface{}{"token": token})
}

/* 获取用户信息 */
func (a *Auth) ParseToken(ctx context.Context, req *modules.RequestPayload) modules.ResponsePayload {
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
