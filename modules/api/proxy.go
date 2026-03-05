package api

import (
	"github.com/gloopai/gloop/modules/auth"
	"github.com/gloopai/gloop/modules/pkg"
)

type Proxy struct {
	Rest    *Api
	Auth    *auth.Auth
	Options *ApiOptions
	Mysql   *pkg.MysqlClient
}

func NewProxy(api *Api) *Proxy {
	return &Proxy{
		Rest:    api,
		Auth:    api.Auth,
		Options: &api.Config,
		Mysql:   api.proxy.Mysql,
	}
}
