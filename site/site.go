package site

import (
	"bytes"
	"crypto/tls"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Site 代表一个具有可配置域和设置的 Web 服务器
type Site struct {
	Config     SiteConfig     // 站点配置
	mux        *http.ServeMux // HTTP 路由器
	JWTManager *JWTManager    // JWT 管理器
	Logger     *log.Logger    // 日志记录器
}

// SiteConfig 保存 Site 的配置
type SiteConfig struct {
	Port       int        `json:"port"`        // 端口号
	TLSCert    string     `json:"tls_cert"`    // cert 证书路径，UseHTTPS 为 true 时需要
	TLSKey     string     `json:"tls_key"`     // key 证书路径，UseHTTPS 为 true 时需要
	UseHTTPS   bool       `json:"use_https"`   // 是否使用 HTTPS
	BaseRoot   string     `json:"base_root"`   // 基础目录
	JWTOptions JWTOptions `json:"jwt_options"` // JWT 选项
	UseEmbed   bool       `json:"use_embed"`   // 是否使用嵌入文件
	EmbedFiles embed.FS   `json:"embed_files"` // 嵌入文件系统
}

// 初始化日志记录器
func NewSite(config SiteConfig) *Site {
	site := &Site{
		Config: config,
		Logger: log.New(os.Stdout, "[Site] ", log.LstdFlags),
	}
	// token 默认值
	if config.JWTOptions.Authorization == "" {
		config.JWTOptions.Authorization = "Authorization"
	}
	site.JWTManager = NewJWTManager(config.JWTOptions)
	return site
}

// 修改 Start 方法以在 Site 级别初始化 mux
func (s *Site) Start() error {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.serveStaticFiles(w, r)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Config.Port),
		Handler: s.mux,
	}

	if s.Config.UseHTTPS {
		if s.Config.TLSCert == "" || s.Config.TLSKey == "" {
			return fmt.Errorf("必须提供 TLS 证书和密钥以启用 HTTPS (端口: %d)", s.Config.Port)
		}

		cert, err := tls.LoadX509KeyPair(s.Config.TLSCert, s.Config.TLSKey)
		if err != nil {
			return fmt.Errorf("加载 TLS 证书和密钥失败 (证书: %s, 密钥: %s): %v", s.Config.TLSCert, s.Config.TLSKey, err)
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		fmt.Printf("启动 HTTPS 服务器，端口 %d \n", s.Config.Port)
		go func() {
			if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTPS 服务器错误: %v\n", err)
			}
		}()
	} else {
		fmt.Printf("启动 HTTP 服务器，端口 %d \n", s.Config.Port)
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTP 服务器错误: %v\n", err)
			}
		}()
	}

	return nil
}

// 处理静态文件
func (s *Site) serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	if s.Config.UseEmbed {
		if err := s.serveEmbeddedFile(w, r); err != nil {
			s.Logger.Printf("嵌入文件服务错误: %v", err)
			http.NotFound(w, r)
		}
		return
	}

	// 回退到从文件系统服务文件
	requestedPath := filepath.Join(s.Config.BaseRoot, r.URL.Path)
	if _, err := os.Stat(requestedPath); err == nil {
		http.ServeFile(w, r, requestedPath)
		return
	}

	// 如果文件不存在，为 React 路由服务 index.html
	filePath := filepath.Join(s.Config.BaseRoot, "index.html")
	file, err := os.Open(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	http.ServeFile(w, r, filePath)
}

// 服务嵌入文件
func (s *Site) serveEmbeddedFile(w http.ResponseWriter, r *http.Request) error {
	requestedPath := filepath.Join(s.Config.BaseRoot, r.URL.Path)
	file, err := s.Config.EmbedFiles.Open(requestedPath)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	http.ServeContent(w, r, requestedPath, time.Now(), bytes.NewReader(content))
	return nil
}

// 注册一个普通路由
func (s *Site) AddRoute(pattern string, handlerFunc http.HandlerFunc) {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}
	s.mux.HandleFunc(pattern, handlerFunc)
}

// 注册一个 payload 路由
func (s *Site) AddPayloadRoute(pattern string, handlerFunc func(payload RequestPayload) ResponsePayload) {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusMethodNotAllowed,
				Message: "方法不被允许",
			})
			return
		}

		// 解析 JSON 请求体
		var payload RequestPayload
		if err := ParseJSONRequest(r, &payload); err != nil {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusBadRequest,
				Message: "无效的 JSON 请求体",
			})
			return
		}

		// 调用处理函数并获取响应
		response := handlerFunc(payload)
		// 将响应写为 JSON
		WriteJSONResponse(w, response)
	})
}

/* 注册一个带 token 验证的路由 */
func (s *Site) AddTokenPayloadRoute(pattern string, handlerFunc func(payload RequestPayload) ResponsePayload) {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusMethodNotAllowed,
				Message: "方法不被允许",
			})
			return
		}

		// 从 Authorization 头中提取 JWT token
		token := r.Header.Get("Authorization")
		if token == "" {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusUnauthorized,
				Message: "缺少 Authorization 头",
			})
			return
		}

		// 验证 token
		auth, err := s.JWTManager.VerifyToken(token)
		if err != nil {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusUnauthorized,
				Message: "无效的 token",
			})
			return
		}

		// 解析 JSON 请求体
		var payload RequestPayload
		if err := ParseJSONRequest(r, &payload); err != nil {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusBadRequest,
				Message: "无效的 JSON 请求体",
			})
			return
		}

		payload.Auth = auth

		// 调用处理函数并获取响应
		response := handlerFunc(payload)
		// 将响应写为 JSON
		WriteJSONResponse(w, response)
	})
}
