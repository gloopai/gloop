package site

import (
	"github.com/gloopai/gloop/modules/auth"
	"github.com/gloopai/gloop/modules/db"
)

type Proxy struct {
	Site      *Site
	Auth      *auth.Auth
	DbService *db.DbService
}

func NewProxy(site *Site) *Proxy {
	return &Proxy{
		Site:      site,
		Auth:      site.Auth,
		DbService: site.DbService,
	}
}
