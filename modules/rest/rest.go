package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gloopai/gloop/component"
	"github.com/gloopai/gloop/lib"
	"github.com/gloopai/gloop/modules/auth"
	"github.com/gloopai/gloop/schema"
)

type Rest struct {
	component.Base
	Config RestOptions    // 站点配置
	mux    *http.ServeMux // HTTP 路由器

	// 在 Rest 结构中添加 RouteCommandMap
	RouteCommandMap *RouteCommandManager
	Auth            *auth.Auth
	proxy           *Proxy
}

// 初始化日志记录器
func NewRest(proxy *Proxy) *Rest {
	auth := auth.NewAuth(
		&auth.Proxy{
			Options: auth.AuthOptions{
				TelegramBotToken: proxy.Options.Auth.TelegramBotToken,
				JWTOptions: auth.JWTOptions{
					SecretKey:     proxy.Options.Auth.Jwt.SecretKey,
					TokenDuration: proxy.Options.Auth.Jwt.TokenDuration,
				},
			},
			Mysql: proxy.Mysql,
		})
	return &Rest{
		Config:          *proxy.Options,
		RouteCommandMap: NewRouteCommandManager(),
		Auth:            auth,
		proxy:           proxy,
	}
}

func (s *Rest) Name() string {
	return "rest"
}

func (s *Rest) Init() {
	if s.Config.Id == "" {
		s.Config.Id = lib.Generate.Guid()
	}

	// s.Auth.SetEnv(s.GetEnv())
	s.Auth.Init() // 初始化 auth 模块

	// s.printInfo()

	if s.mux == nil {
		s.mux = http.NewServeMux()
	}
}

// 修改 Start 方法以在 Rest 级别初始化 mux
func (s *Rest) Start() error {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	// 优化 HTTP 服务器配置
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", s.Config.Port),
		Handler:        s.mux,
		ReadTimeout:    10 * time.Second, // 限制读取超时时间
		WriteTimeout:   10 * time.Second, // 限制写入超时时间
		MaxHeaderBytes: 1 << 20,          // 限制请求头大小为 1MB
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP 服务器错误: %v\n", err)
		}
	}()

	return nil
}

// AddRoute 注册一个普通的 HTTP 路由
//
// 该方法接受一个 URL 模式和一个处理函数，并将它们添加到 HTTP 服务器的路由器中。
// 当客户端发送请求到指定路径时，服务器将调用相应的处理函数来处理该请求。
//
// 参数：
//   - pattern: URL 路径模式，例如 "/hello" 或 "/api/users"
//   - handlerFunc: 处理该路由的函数
//
// 示例：
//
//	rest.AddRoute("/hello", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Hello, World!"))
//	})
func (s *Rest) AddRoute(pattern string, handlerFunc http.HandlerFunc) {
	defer func() {
		if r := recover(); r != nil {
			lib.Log.Errorf("AddRoute panic: %v\n", r)
		}
	}()
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}
	s.mux.HandleFunc(pattern, handlerFunc)
}

// AddPayloadRoute 注册一个处理 JSON 请求体的路由
//
// 该方法创建一个路由来处理带有 JSON 请求体的 POST 请求。
// 请求体应该包含 command 字段来指定要执行的命令。
//
// 请求格式：
//
//	{
//	    "command": "example_command",
//	    "data": {}
//	}
//
// 响应格式：
//
//	{
//	    "code": 200,
//	    "message": "success",
//	    "data": {},
//	    "traceId": "uuid"
//	}
func (s *Rest) AddPayloadRoute(pattern string) {
	defer func() {
		if r := recover(); r != nil {
			lib.Log.Errorf("AddPayloadRoute panic: %v\n", r)
		}
	}()
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		s.handlePayloadRequest(w, r, pattern, nil)
	})
}

// AddPayloadRouteWithAuth 注册一个需要身份验证的 JSON 请求体路由
//
// 该方法类似于 AddPayloadRoute，但要求请求包含有效的 Authorization 头。
// JWT token 将被验证，用户信息将被添加到请求上下文中。
//
// Headers:
//
//	Authorization: Bearer <jwt_token>
//
// 请求格式和响应格式与 AddPayloadRoute 相同。
// 如果身份验证失败，将返回 401 状态码。
func (s *Rest) AddPayloadRouteWithAuth(pattern string) {
	defer func() {
		if r := recover(); r != nil {
			lib.Log.Errorf("AddPayloadRouteWithAuth panic: %v\n", r)
		}
	}()
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if s.Auth == nil {
			schema.WriteJSONResponse(w, schema.Response{
				Code:    http.StatusInternalServerError,
				Message: "Auth module not initialized",
			})
			return
		}
		// 从 Authorization 头中提取 JWT token
		token := r.Header.Get(s.Auth.Authorization())
		if token == "" {
			schema.WriteJSONResponse(w, schema.Response{
				Code:    http.StatusUnauthorized,
				Message: "Missing Authorization header",
			})
			return
		}

		// 验证 token
		auth, err := s.Auth.JWTManager.VerifyToken(token)
		if err != nil {
			schema.WriteJSONResponse(w, schema.Response{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token 1 " + err.Error(),
			})
			return
		}

		if auth.UserId == 0 {
			schema.WriteJSONResponse(w, schema.Response{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token 2",
			})
			return
		}

		s.handlePayloadRequest(w, r, pattern, &auth)
	})
}

// handlePayloadRequest 处理 JSON 请求体的公共逻辑
//
// 该辅助函数处理 JSON 请求的解析、验证和路由分发。
// 它设置 CORS 头，解析请求体，并根据 command 字段调用相应的处理器。
func (s *Rest) handlePayloadRequest(w http.ResponseWriter, r *http.Request, pattern string, auth *schema.RequestAuth) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// 解析 JSON 请求体
	var payload schema.Request
	traceID := lib.Generate.Guid()
	r = r.WithContext(context.WithValue(r.Context(), schema.TraceIDContextKey, traceID))
	payload.TraceId = traceID
	w.Header().Set("X-Trace-Id", traceID)

	// 处理预检请求
	if r.Method != http.MethodPost {
		schema.WriteJSONResponse(w, schema.Response{
			TraceId: traceID,
			Code:    http.StatusMethodNotAllowed,
			Message: "Method not allowed",
		})
		return
	}

	if err := schema.ParseJSONRequest(r, &payload); err != nil {
		schema.WriteJSONResponse(w, schema.Response{
			TraceId: traceID,
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
		response := handler(r.Context(), &payload)
		response.TraceId = traceID
		schema.WriteJSONResponse(w, response)
		return
	}

	schema.WriteJSONResponse(w, schema.Response{
		Code:    http.StatusNotFound,
		TraceId: traceID,
		Message: "Command not found",
	})
}

// RegisterCommand 向通过 AddRoute 方法添加的路由注册一个命令处理函数
// 请求体应该包含 command 字段来指定要执行的命令。
//
// 请求格式：
//
//	{
//	    "command": "example_command",
//	    "data": {}
//	}
//
// 响应格式：
//
//	{
//	    "code": 200,
//	    "message": "success",
//	    "data": {},
//	    "traceId": "uuid"
//	}

func (s *Rest) RegisterCommand(route string, command string, handler func(ctx context.Context, payload *schema.Request) schema.Response) {
	key := fmt.Sprintf("%s:%s", route, command)
	s.RouteCommandMap.Store(key, handler)
}

// GetBindAddresses 返回当前 HTTP 服务器绑定的 IP 和端口列表
func (s *Rest) GetBindAddresses() []string {
	addresses := make([]string, 0)

	protocol := "http"
	// 构建绑定地址，如果绑定到所有接口（端口前缀为空或 ':'），则返回常见接口
	address := fmt.Sprintf(":%d", s.Config.Port)

	// 绑定到所有接口
	if address[0] == ':' {
		addresses = append(addresses, fmt.Sprintf("%s://0.0.0.0%s", protocol, address))
		addresses = append(addresses, fmt.Sprintf("%s://[::]%s", protocol, address))
		addresses = append(addresses, fmt.Sprintf("%s://localhost%s", protocol, address))
	} else {
		// 绑定到特定地址
		addresses = append(addresses, fmt.Sprintf("%s://%s", protocol, address))
	}

	return addresses
}

func (s *Rest) Proxy() *Proxy {
	return &Proxy{
		Rest: s,
		Auth: s.Auth,
	}
}
