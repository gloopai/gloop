package auth

import "github.com/gloopai/gloop/modules/pkg"

type Proxy struct {
	Options AuthOptions
	Mysql   *pkg.MysqlClient
}

func NewProxy(opt AuthOptions) *Proxy {
	return &Proxy{
		Options: opt,
	}
}

func (p *Proxy) GetMysql() (*pkg.MysqlClient, error) {
	return p.Mysql, nil
}
