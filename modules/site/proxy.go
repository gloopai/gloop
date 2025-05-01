package site

import "github.com/gloopai/gloop/modules/db"

type Proxy struct {
	Site *Site
	Db   *db.DbSqlite
}

func NewProxy(site *Site) *Proxy {
	return &Proxy{
		Site: site,
	}
}
