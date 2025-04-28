package site

import (
	"fmt"
	"net/http"
)

type SiteServer struct {
	Debug   bool
	BaseURL string
	Port    int
}

type SiteServerOptions struct {
	Debug   bool
	BaseURL string
	Port    int
}

func NewSiteServer(configPath string) *SiteServer {
	return &SiteServer{}
}

func (s *SiteServer) Start() {
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", 9999), nil)
}
