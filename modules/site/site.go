package site

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules"
	"github.com/gloopai/gloop/modules/auth"
)

// Site 代表一个具有可配置域和设置的 Web 服务器
type Site struct {
	modules.Base
	Config SiteOptions    // 站点配置
	mux    *http.ServeMux // HTTP 路由器

	// 在 Site 结构中添加 RouteCommandMap
	RouteCommandMap *RouteCommandManager
	Auth            *auth.Auth
}

// 初始化日志记录器
func NewSite(config SiteOptions) *Site {
	return &Site{
		Config:          config,
		RouteCommandMap: NewRouteCommandManager(),
	}
}

func (s *Site) Name() string {
	return "site"
}

func (s *Site) Init() {
	if s.Config.Id == "" {
		s.Config.Id = lib.Generate.Guid()
	}

	// 在 NewSite 函数中设置 StaticFileCacheTTL 的默认值
	if s.Config.StaticFileCacheTTL == 0 {
		s.Config.StaticFileCacheTTL = 10 * time.Minute // 默认值为 10 分钟
	}

	s.printInfo()
}

func (s *Site) Close() {}

func (s *Site) Destory() {}

// // 打印组件信息
func (s *Site) printInfo() {
	infos := make([]string, 0, 7)
	infos = append(infos, fmt.Sprintf("ID: %s", s.Config.Id))
	// infos = append(infos, fmt.Sprintf("Name: %s", s.Name()))
	infos = append(infos, fmt.Sprintf("Port: %d", s.Config.Port))
	infos = append(infos, fmt.Sprintf("UseEmbed: %t", s.Config.UseEmbed))
	infos = append(infos, fmt.Sprintf("BaseRoot: %s", s.Config.BaseRoot))
	infos = append(infos, fmt.Sprintf("UseHTTPS: %t", s.Config.UseHTTPS))
	if s.Config.UseHTTPS {
		infos = append(infos, fmt.Sprintf("TLSCert: %s", s.Config.TLSCert))
		infos = append(infos, fmt.Sprintf("TLSKey: %s", s.Config.TLSKey))
	}

	infos = append(infos, fmt.Sprintf("ForceIndexHTML: %t", s.Config.ForceIndexHTML))
	infos = append(infos, fmt.Sprintf("StaticFileCacheTTL: %s", s.Config.StaticFileCacheTTL))

	modules.PrintBoxInfo(s.Name(), infos...)
}

// 修改 Start 方法以在 Site 级别初始化 mux
func (s *Site) Start() error {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.serveStaticFiles(w, r)
	})

	// 优化 HTTP 服务器配置
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", s.Config.Port),
		Handler:        s.mux,
		ReadTimeout:    10 * time.Second, // 限制读取超时时间
		WriteTimeout:   10 * time.Second, // 限制写入超时时间
		MaxHeaderBytes: 1 << 20,          // 限制请求头大小为 1MB
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

		go func() {
			if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTPS 服务器错误: %v\n", err)
			}
		}()
	} else {
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTP 服务器错误: %v\n", err)
			}
		}()
	}

	return nil
}

// 替换 serveStaticFiles 调用为使用 StaticFileHandler
func (s *Site) serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	config := StaticFileHandlerConfig{
		TTL:            s.Config.StaticFileCacheTTL,
		BaseRoot:       s.Config.BaseRoot,
		UseEmbed:       s.Config.UseEmbed,
		EmbedFS:        s.Config.EmbedFiles,
		ForceIndexHTML: s.Config.ForceIndexHTML,
	}
	staticFileHandler := NewStaticFileHandler(config)
	staticFileHandler.StartCacheCleaner()
	staticFileHandler.ServeStaticFile(w, r)
}

// 注册一个普通路由
func (s *Site) AddRoute(pattern string, handlerFunc http.HandlerFunc) {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}
	s.mux.HandleFunc(pattern, handlerFunc)
}

// 修改 RegisterCommand 方法以适配 sync.Map
func (s *Site) RegisterPayloadCommand(route string, command string, handler func(*modules.RequestPayload) modules.ResponsePayload) {
	key := fmt.Sprintf("%s:%s", route, command)
	s.RouteCommandMap.Store(key, handler)
}

// 修改 AddPayloadRoute 方法以适配 sync.Map
func (s *Site) AddPayloadRoute(pattern string) {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		s.handlePayloadRequest(w, r, pattern, nil)
	})
}

// 提取公共逻辑到辅助函数
func (s *Site) handlePayloadRequest(w http.ResponseWriter, r *http.Request, pattern string, auth *modules.RequestAuth) {
	if r.Method != http.MethodPost {
		modules.WriteJSONResponse(w, modules.ResponsePayload{
			Code:    http.StatusMethodNotAllowed,
			Message: "Method not allowed",
		})
		return
	}

	// 解析 JSON 请求体
	var payload modules.RequestPayload
	if err := modules.ParseJSONRequest(r, &payload); err != nil {
		modules.WriteJSONResponse(w, modules.ResponsePayload{
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
		response := handler(&payload)
		modules.WriteJSONResponse(w, response)
		return
	}

	modules.WriteJSONResponse(w, modules.ResponsePayload{
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
		if s.Auth == nil {
			modules.WriteJSONResponse(w, modules.ResponsePayload{
				Code:    http.StatusInternalServerError,
				Message: "Auth module not initialized",
			})
			return
		}
		// 从 Authorization 头中提取 JWT token
		token := r.Header.Get(s.Auth.Authorization())
		if token == "" {
			modules.WriteJSONResponse(w, modules.ResponsePayload{
				Code:    http.StatusUnauthorized,
				Message: "Missing Authorization header",
			})
			return
		}

		// 验证 token
		auth, err := s.Auth.JWTManager.VerifyToken(token)
		if err != nil {
			modules.WriteJSONResponse(w, modules.ResponsePayload{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token",
			})
			return
		}

		s.handlePayloadRequest(w, r, pattern, &auth)
	})
}

/* 使用 auth 模块 */
func (s *Site) UseAuth(auth *auth.Auth) {
	s.Auth = auth
}
