package auth

import (
	"context"
	"fmt"

	"github.com/gloopai/gloop/component"
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/schema"
)

type Auth struct {
	component.Base
	Config     AuthOptions // 认证配置
	JWTManager *JWTManager // JWT 管理器
	proxy      *Proxy      // 代理对象
}

func NewAuth(proxy *Proxy) *Auth {
	return &Auth{
		Config: proxy.Options,
		proxy:  proxy,
	}
}
func (a *Auth) Name() string {
	return "auth"
}
func (a *Auth) Init() {

	if a.Config.JWTOptions.Authorization == "" {
		a.Config.JWTOptions.Authorization = "Authorization"
	}

	a.JWTManager = NewJWTManager(a.Config.JWTOptions)

}

func (a *Auth) Start() error {
	conn, err := a.proxy.GetMysql()
	if err != nil {
		return fmt.Errorf("failed to get mysql client: %v", err)
	}
	err = EnsureAuthTableExists(conn.Conn)
	if err != nil {
		lib.Log.Error("Failed to ensure auth table exists:", err)
	}
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
func (a *Auth) Register(ctx context.Context, req *schema.Request) schema.Response {
	type queryObj struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	var query queryObj
	err := req.Unmarshal(&query)
	if err != nil {
		return schema.PayloadResponse.Error(err.Error())
	}
	conn, err := a.proxy.GetMysql()
	if err != nil {
		return schema.PayloadResponse.Error(err.Error())
	}
	err = RegisterUser(conn.Conn, query.Username, query.Password, query.Email)
	if err != nil {
		return schema.PayloadResponse.Error(err.Error())
	}

	return schema.PayloadResponse.SuccessNone()
}

/* 用户登录 */
func (a *Auth) Login(ctx context.Context, req *schema.Request) schema.Response {
	type queryObject struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var query queryObject
	err := req.Unmarshal(&query)
	if err != nil {
		return schema.PayloadResponse.Error(err.Error())
	}
	conn, err := a.proxy.GetMysql()
	if err != nil {
		return schema.PayloadResponse.Error(err.Error())
	}

	loggedInUser, err := LoginUser(conn.Conn, query.Username, query.Password)
	if err != nil {
		return schema.PayloadResponse.Error(err.Error())
	}

	token, err := a.JWTManager.GenerateToken(schema.RequestAuth{
		UserId:   loggedInUser.Id,
		Username: loggedInUser.Username,
	})
	// Populate the user details
	if err != nil {
		return schema.PayloadResponse.Error(err.Error())
	}

	// loggedInUser.Token = token
	resmap := make(map[string]interface{})
	resmap["token"] = token

	return schema.PayloadResponse.Success(resmap)
}

// 通过 Telegram 登录
func (a *Auth) LoginByTelegram(ctx context.Context, req *schema.Request) schema.Response {
	type queryObject struct {
		InitData string `json:"init_data"`
	}
	var query queryObject
	err := req.Unmarshal(&query)
	if err != nil {
		return schema.PayloadResponse.Error(err.Error())
	}
	telegramUser := &TelegramUser{}
	conn, err := a.proxy.GetMysql()
	if err != nil {
		return schema.PayloadResponse.Error(err.Error())
	}
	_, err = telegramUser.Login(conn.Conn, query.InitData, a.Config.TelegramBotToken)
	if err != nil {
		return schema.PayloadResponse.Error(fmt.Sprintf("Telegram parse error: %s", err.Error()))
	}

	token, err := a.JWTManager.GenerateToken(schema.RequestAuth{
		UserId:   telegramUser.UserId,
		Username: "", // TelegramUser struct does not have a Username field, you might want to add it or handle differently
	})

	// telegramUser, err := LoginUserByTelegram(a.db.Db, query.InitData)
	return schema.PayloadResponse.Success(map[string]interface{}{"token": token})
}

/* 获取用户信息 */
func (a *Auth) ParseToken(ctx context.Context, req *schema.Request) schema.Response {
	type queryObject struct {
		Token string `json:"token"`
	}
	var query queryObject
	err := req.Unmarshal(&query)
	if err != nil {
		return schema.PayloadResponse.Error(err.Error())
	}
	auth, err := a.JWTManager.VerifyToken(query.Token)
	if err != nil {
		return schema.PayloadResponse.Error(fmt.Sprintf("JWTERROR:%s", err.Error()))
	}

	return schema.PayloadResponse.Success(auth)
}
