package base

import (
	"github.com/itpkg/web"
	"github.com/op/go-logging"
)

type Engine struct {
	Mux    *web.Mux        `inject:""`
	Logger *logging.Logger `inject:""`
}

func (p *Engine) Migrate() error {
	return nil
}

func (p *Engine) Seed() error {
	return nil
}

func (p *Engine) Mount() {
	router := web.NewRouter()
	router.GET("^/sitemap.xml.gz$", p.Sitemap)
	router.GET("^/rss.atom$", p.Rss)

	p.Mux.AddRouter(router)
}

func (p Engine) Info() (string, string) {
	return "base", "v20150917"
}

//-----------------------------------------------------------------------------
func init() {
	web.Register(&Engine{})
}
