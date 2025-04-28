package site

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

}
