package auth

import "github.com/gloopai/gloop/modules/db"

type AuthOptions struct {
	Db         *db.DbService
	JWTOptions JWTOptions `json:"jwt_options"` // JWT 选项

	TelegramBotToken string `json:"telegram_bot_token"` // 是否启用 telegram 机器人登录
}
