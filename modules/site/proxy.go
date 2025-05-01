package site

import "github.com/gloopai/gloop/modules/db"

type Proxy struct {
	Site *Site
	Db   *db.Db
}

func NewProxy(site *Site) *Proxy {
	return &Proxy{
		Site: site,
	}
}
