package site

type Site struct {
}

func NewSite() *Site {
	return &Site{}
}

func (s *Site) New() error {
	return nil
}
