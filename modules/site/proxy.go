package site

import (
	"github.com/gloopai/gloop/modules/auth"
)

type Proxy struct {
	Site *Site
	Auth *auth.Auth
}

func NewProxy(site *Site) *Proxy {
	return &Proxy{
		Site: site,
		Auth: site.Auth,
	}
}
