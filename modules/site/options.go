package site

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gloopai/gloop/lib"
)

// SiteConfig 保存 Site 的配置
type SiteOptions struct {
	Id             string   `json:"id"`               // 站点 ID
	Port           int      `json:"port"`             // 端口号
	TLSCert        string   `json:"tls_cert"`         // cert 证书路径，UseHTTPS 为 true 时需要
	TLSKey         string   `json:"tls_key"`          // key 证书路径，UseHTTPS 为 true 时需要
	UseHTTPS       bool     `json:"use_https"`        // 是否使用 HTTPS
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
		Id:             lib.Generate.Guid(),
		Port:           8080,
		TLSCert:        "",
		TLSKey:         "",
		UseHTTPS:       false,
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
		TLSCert:        options.TLSCert,
		TLSKey:         options.TLSKey,
		UseHTTPS:       options.UseHTTPS,
		BaseRoot:       options.BaseRoot,
		UseEmbed:       false,
		EmbedFiles:     embed.FS{},
		ForceIndexHTML: true,
	}
}

func LoadSiteTOMLOptions(path string) SiteOptions {
	tomlData, err := loadFileToBytes(path)
	if err != nil {
		fmt.Println("Error loading file:", err)
		return DefaultOptions()
	}
	var options SiteOptions

	// 解析 TOML 数据
	_, err = toml.Decode(string(tomlData), &options)
	if err != nil {
		fmt.Println("Error decoding TOML:", err)
		return DefaultOptions()
	}

	return SiteOptions{
		Id:             options.Id,
		Port:           options.Port,
		TLSCert:        options.TLSCert,
		TLSKey:         options.TLSKey,
		UseHTTPS:       options.UseHTTPS,
		BaseRoot:       options.BaseRoot,
		UseEmbed:       false,
		EmbedFiles:     embed.FS{},
		ForceIndexHTML: true,
	}
}

func loadFileToBytes(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)

	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}
