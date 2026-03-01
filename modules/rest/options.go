package rest

import (
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules/auth"
)

type SiteAuthOption struct {
	TelegramBotToken string          `json:"telegram_bot_token"` // Telegram Bot Token
	Jwt              auth.JWTOptions `json:"jwt"`                // JWT
}

// RestOptions 保存 Rest 的配置
type RestOptions struct {
	Id   string `json:"id"`   // 站点 ID
	Port int    `json:"port"` // 端口号
	// 在 RestOptions 中添加 CrossOrigin 配置项
	CrossOrigin bool `json:"cross_origin"` // 是否启用跨域
	/// Auth 模块配置
	Auth SiteAuthOption `json:"auth"`
}

func DefaultOptions() RestOptions {
	return RestOptions{
		Id:   lib.Generate.Guid(),
		Port: 8080,
		Auth: SiteAuthOption{
			TelegramBotToken: "",
			Jwt: auth.JWTOptions{
				SecretKey:     "RxyiJcD8O19/GE9GL/V2sn0b/MOSWTWoygN77e7RNSI=",
				TokenDuration: 24 * 365, // 默认 token 有效期为 24 小时
			},
		},
	}

}
