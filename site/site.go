package site

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// Site represents a web server with configurable domains and settings.
type Site struct {
	Config     SiteConfig     // Configuration for the site
	mux        *http.ServeMux // HTTP request multiplexer
	JWTManager *JWTManager    // JWT manager for token handling
}

// SiteConfig holds the configuration for the Site.
type SiteConfig struct {
	Port     int    `json:"port"`      // Port to run the server on
	TLSCert  string `json:"tls_cert"`  // Path to the TLS certificate
	TLSKey   string `json:"tls_key"`   // Path to the TLS key
	UseHTTPS bool   `json:"use_https"` // Whether to use HTTPS

	BaseRoot   string     `json:"base_root"`   // Base directory for static files
	JWTOptions JWTOptions `json:"jwt_options"` // Secret key for JWT
}

func NewSite(config SiteConfig) *Site {
	site := &Site{
		Config: config,
	}
	site.JWTManager = NewJWTManager(config.JWTOptions)
	return site
}

// Modify the Start method to initialize the mux at the Site level
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
			return fmt.Errorf("TLS certificate and key must be provided for HTTPS (Port: %d)", s.Config.Port)
		}

		cert, err := tls.LoadX509KeyPair(s.Config.TLSCert, s.Config.TLSKey)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificate and key (Cert: %s, Key: %s): %v", s.Config.TLSCert, s.Config.TLSKey, err)
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		fmt.Printf("Starting HTTPS server on port %d \n", s.Config.Port)
		go func() {
			if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTPS server error: %v\n", err)
			}
		}()
	} else {
		fmt.Printf("Starting HTTP server on port %d \n", s.Config.Port)
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		}()
	}

	return nil
}

// 处理静态文件
func (s *Site) serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	requestedPath := filepath.Join(s.Config.BaseRoot)

	if _, err := os.Stat(requestedPath); err == nil {
		http.ServeFile(w, r, requestedPath)
		return
	}

	// If the file does not exist, serve index.html for React routing
	filePath := requestedPath + "/index.html"
	file, err := os.Open(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	http.ServeFile(w, r, filePath)
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
				Message: "Method not allowed",
			})
			return
		}

		// Parse the JSON request body
		var payload RequestPayload
		if err := ParseJSONRequest(r, &payload); err != nil {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusBadRequest,
				Message: "Invalid JSON payload",
			})
			return
		}

		// Call the handler function and get the response
		response := handlerFunc(payload)
		// Write the response as JSON
		WriteJSONResponse(w, response)
	})
}

/* 注册一个 token 验证的路由 */
func (s *Site) AddTokenPayloadRoute(pattern string, handlerFunc func(payload RequestPayload) ResponsePayload) {
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

		// Extract JWT token from Authorization header
		token := r.Header.Get("Authorization")
		if token == "" {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusUnauthorized,
				Message: "Authorization header missing",
			})
			return
		}

		// Verify the token
		auth, err := s.JWTManager.VerifyToken(token)
		if err != nil {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token",
			})
			return
		}

		// Parse the JSON request body
		var payload RequestPayload
		if err := ParseJSONRequest(r, &payload); err != nil {
			WriteJSONResponse(w, ResponsePayload{
				Code:    http.StatusBadRequest,
				Message: "Invalid JSON payload",
			})
			return
		}

		payload.Auth = auth

		// Call the handler function and get the response
		response := handlerFunc(payload)
		// Write the response as JSON
		WriteJSONResponse(w, response)
	})
}
