package base

import (
	"github.com/itpkg/web"
)

func (p *Engine) Sitemap(c *web.Context) *web.HttpError {
	//todo
	return nil
}

func (p *Engine) Rss(c *web.Context) *web.HttpError {
	//todo
	return nil
}

func (p *Engine) Locales(c *web.Context) *web.HttpError {
	var items []Locale
	p.Db.Select([]string{"code", "message"}).Where("lang = ?", c.Params["locale"]).Order("code DESC").Find(&items)
	return c.JSON(&items)
}
