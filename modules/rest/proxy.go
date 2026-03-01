package rest

import (
	"github.com/gloopai/gloop/modules/auth"
	"github.com/gloopai/gloop/modules/pkg"
)

type Proxy struct {
	Rest    *Rest
	Auth    *auth.Auth
	Options *RestOptions
	Mysql   *pkg.MysqlClient
}

func NewProxy(rest *Rest) *Proxy {
	return &Proxy{
		Rest:    rest,
		Auth:    rest.Auth,
		Options: &rest.Config,
		Mysql:   rest.proxy.Mysql,
	}
}
