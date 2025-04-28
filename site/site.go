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
	Domains []string       // List of domains the site serves
	Config  SiteConfig     // Configuration for the site
	mux     *http.ServeMux // HTTP request multiplexer
}

// SiteConfig holds the configuration for the Site.
type SiteConfig struct {
	Port     int    `json:"port"`      // Port to run the server on
	TLSCert  string `json:"tls_cert"`  // Path to the TLS certificate
	TLSKey   string `json:"tls_key"`   // Path to the TLS key
	UseHTTPS bool   `json:"use_https"` // Whether to use HTTPS
	BaseRoot string `json:"base_root"` // Base directory for static files
}

func NewSite(config SiteConfig, domains []string) *Site {
	return &Site{
		Domains: domains,
		Config:  config,
	}
}

// Modify the Start method to initialize the mux at the Site level
func (s *Site) Start() error {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}

	// Map each domain to its corresponding static HTML folder
	for _, domain := range s.Domains {
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			s.serveStaticFiles(domain, w, r)
		})
	}

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

		fmt.Printf("Starting HTTPS server on port %d for domains: %v\n", s.Config.Port, s.Domains)
		go func() {
			if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTPS server error: %v\n", err)
			}
		}()
	} else {
		fmt.Printf("Starting HTTP server on port %d for domains: %v\n", s.Config.Port, s.Domains)
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		}()
	}

	return nil
}

// serveStaticFiles serves static files from the specified directory.
func (s *Site) serveStaticFiles(domain string, w http.ResponseWriter, r *http.Request) {
	requestedPath := filepath.Join(s.Config.BaseRoot, domain)

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

// Update RegisterRoute to use the Site-level mux
func (s *Site) RegisterRoute(pattern string, handlerFunc http.HandlerFunc) {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}
	s.mux.HandleFunc(pattern, handlerFunc)
}
