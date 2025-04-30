package site

import (
	"embed"
	"time"
)

// SiteConfig 保存 Site 的配置
type SiteOptions struct {
	Id             string     `json:"id"`               // 站点 ID
	Port           int        `json:"port"`             // 端口号
	TLSCert        string     `json:"tls_cert"`         // cert 证书路径，UseHTTPS 为 true 时需要
	TLSKey         string     `json:"tls_key"`          // key 证书路径，UseHTTPS 为 true 时需要
	UseHTTPS       bool       `json:"use_https"`        // 是否使用 HTTPS
	BaseRoot       string     `json:"base_root"`        // 基础目录
	JWTOptions     JWTOptions `json:"jwt_options"`      // JWT 选项
	UseEmbed       bool       `json:"use_embed"`        // 是否使用嵌入文件
	EmbedFiles     embed.FS   `json:"embed_files"`      // 嵌入文件系统
	ForceIndexHTML bool       `json:"force_index_html"` // 是否强制使用 index.html

	// 在 SiteConfig 中添加 StaticFileCacheTTL 配置项
	StaticFileCacheTTL time.Duration `json:"static_file_cache_ttl"`
}
