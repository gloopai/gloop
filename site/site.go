package site

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type Site struct {
	Domains []string
	Config  SiteConfig
}

type SiteConfig struct {
	Port     int    `json:"port"`
	TLSCert  string `json:"tls_cert"`
	TLSKey   string `json:"tls_key"`
	UseHTTPS bool   `json:"use_https"`
	BaseRoot string `json:"base_root"`
}

func NewSite(config SiteConfig, domains []string) *Site {
	return &Site{
		Domains: domains,
		Config:  config,
	}
}

func (s *Site) Start() error {
	mux := http.NewServeMux()

	// Map each domain to its corresponding static HTML folder
	for _, domain := range s.Domains {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			s.serveStaticFiles(domain, w, r)
		})
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Config.Port),
		Handler: mux,
	}

	if s.Config.UseHTTPS {
		if s.Config.TLSCert == "" || s.Config.TLSKey == "" {
			return fmt.Errorf("TLS certificate and key must be provided for HTTPS")
		}

		cert, err := tls.LoadX509KeyPair(s.Config.TLSCert, s.Config.TLSKey)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificate and key: %v", err)
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		fmt.Printf("Starting HTTPS server on port %d for domains: %v\n", s.Config.Port, s.Domains)
		return server.ListenAndServeTLS("", "")
	}

	fmt.Printf("Starting HTTP server on port %d for domains: %v\n", s.Config.Port, s.Domains)
	return server.ListenAndServe()
}

// serveStaticFiles serves static files from the specified directory
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
