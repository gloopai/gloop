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
	"sync"
	"time"
)

// Site 代表一个具有可配置域和设置的 Web 服务器
type Site struct {
	Config     SiteConfig     // 站点配置
	mux        *http.ServeMux // HTTP 路由器
	JWTManager *JWTManager    // JWT 管理器
	Logger     *log.Logger    // 日志记录器

	// 在 Site 结构中添加 RouteCommandMap
	RouteCommandMap sync.Map
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

	// 初始化 RouteCommandMap
	site.RouteCommandMap = sync.Map{}
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
			s.Logger.Printf("Error serving embedded file: %v", err)
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

// 修改 RegisterCommand 方法以适配 sync.Map
func (s *Site) RegisterCommand(route string, command string, handler func(*RequestPayload) ResponsePayload) {
	key := fmt.Sprintf("%s:%s", route, command)
	s.RouteCommandMap.Store(key, handler)
}

// 修改 AddPayloadRoute 方法以适配 sync.Map
func (s *Site) AddPayloadRoute(pattern string) {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusMethodNotAllowed,
				Message: "Method not allowed",
			})
			return
		}

		// 解析 JSON 请求体
		var payload RequestPayload
		if err := ParseJSONRequest(r, &payload); err != nil {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusBadRequest,
				Message: "Invalid JSON payload",
			})
			return
		}

		// 根据 Command 执行对应的处理函数
		key := fmt.Sprintf("%s:%s", pattern, payload.Command)
		if handler, ok := s.RouteCommandMap.Load(key); ok {
			response := handler.(func(*RequestPayload) ResponsePayload)(&payload)
			WriteJSONResponse(w, response)
			return
		}

		WriteJSONResponse(w, ResponsePayload{
			Code:    http.StatusNotFound,
			Message: "Command not found",
		})
	})
}

// 提取公共逻辑到辅助函数
func (s *Site) handlePayloadRequest(w http.ResponseWriter, r *http.Request, pattern string, auth *RequestAuth) {
	if r.Method != http.MethodPost {
		WriteJSONResponse(w, ResponsePayload{
			Code:    http.StatusMethodNotAllowed,
			Message: "Method not allowed",
		})
		return
	}

	// 解析 JSON 请求体
	var payload RequestPayload
	if err := ParseJSONRequest(r, &payload); err != nil {
		WriteJSONResponse(w, ResponsePayload{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON payload",
		})
		return
	}

	if auth != nil {
		payload.Auth = *auth
	}

	// 根据 Command 执行对应的处理函数
	key := fmt.Sprintf("%s:%s", pattern, payload.Command)
	if handler, ok := s.RouteCommandMap.Load(key); ok {
		response := handler.(func(*RequestPayload) ResponsePayload)(&payload)
		WriteJSONResponse(w, response)
		return
	}

	WriteJSONResponse(w, ResponsePayload{
		Code:    http.StatusNotFound,
		Message: "Command not found",
	})
}

// 修改 AddTokenPayloadRoute 方法以使用辅助函数
func (s *Site) AddTokenPayloadRoute(pattern string) {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// 从 Authorization 头中提取 JWT token
		token := r.Header.Get("Authorization")
		if token == "" {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusUnauthorized,
				Message: "Missing Authorization header",
			})
			return
		}

		// 验证 token
		auth, err := s.JWTManager.VerifyToken(token)
		if err != nil {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token",
			})
			return
		}

		s.handlePayloadRequest(w, r, pattern, &auth)
	})
}

// 生成 JWT token
func (s *Site) GenerateToken(auth RequestAuth) (string, error) {
	token, err := s.JWTManager.GenerateToken(auth)
	if err != nil {
		return "", err
	}
	return token, nil
}
