package site

type Proxy struct {
	Site *Site
}

func NewProxy(site *Site) *Proxy {
	return &Proxy{
		Site: site,
	}
}
