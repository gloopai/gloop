package site

import (
	"github.com/gloopai/gloop/events"
	"github.com/gloopai/gloop/modules"
	"github.com/gloopai/gloop/modules/auth"
)

type Proxy struct {
	Events  *events.EventBus
	Site    *Site
	Auth    *auth.Auth
	Context *modules.ComponentContext
}

func NewProxy(site *Site) *Proxy {
	return &Proxy{
		Events:  site.events,
		Site:    site,
		Auth:    site.Auth,
		Context: site.GetContext(),
	}
}
