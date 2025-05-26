package site

import (
	"embed"
	"encoding/json"
	"os"
	"time"

	"github.com/gloopai/gloop/lib"
)

type SiteCert struct {
	CertFile string `json:"CertFile"` // 证书内容
	KeyFile  string `json:"KeyFile"`  // 密钥内容
}

// SiteConfig 保存 Site 的配置
type SiteOptions struct {
	Id             string   `json:"id"`               // 站点 ID
	Port           int      `json:"port"`             // 端口号
	UseHTTPS       bool     `json:"use_https"`        // 是否使用 HTTPS
	Cert           SiteCert `json:"cert"`             // 证书配置
	BaseRoot       string   `json:"base_root"`        // 基础目录
	UseEmbed       bool     `json:"use_embed"`        // 是否使用嵌入文件
	EmbedFiles     embed.FS `json:"embed_files"`      // 嵌入文件系统
	ForceIndexHTML bool     `json:"force_index_html"` // 是否强制使用 index.html

	// 在 SiteConfig 中添加 StaticFileCacheTTL 配置项
	StaticFileCacheTTL time.Duration `json:"static_file_cache_ttl"`

	// 在 SiteConfig 中添加 CrossOrigin 配置项
	CrossOrigin bool `json:"cross_origin"` // 是否启用跨域
}

func DefaultOptions() SiteOptions {
	return SiteOptions{
		Id:       lib.Generate.Guid(),
		Port:     8080,
		UseHTTPS: false,
		Cert: SiteCert{
			CertFile: "",
			KeyFile:  "",
		},
		BaseRoot:       "./",
		UseEmbed:       false,
		EmbedFiles:     embed.FS{},
		ForceIndexHTML: true,
	}

}

/* 通过 json 读取站点配置 */
func LoadSiteJSONOptions(path string) SiteOptions {
	file, err := os.Open(path)
	if err != nil {
		return DefaultOptions()
	}
	defer file.Close()

	var options SiteOptions
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&options); err != nil {
		return DefaultOptions()
	}

	return SiteOptions{
		Id:             options.Id,
		Port:           options.Port,
		Cert:           options.Cert,
		UseHTTPS:       options.UseHTTPS,
		BaseRoot:       options.BaseRoot,
		UseEmbed:       false,
		EmbedFiles:     embed.FS{},
		ForceIndexHTML: true,
	}
}

func LoadSiteTOMLOptions(path string) SiteOptions {
	var options SiteOptions
	err := lib.Conf.LoadTOML(path, &options)
	if err != nil {
		lib.Log.Error("Error loading TOML file:", err)
		return DefaultOptions()
	}
	return options
}
