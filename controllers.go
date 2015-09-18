package base

import (
	"strings"

	"github.com/itpkg/web"
)

func (p *Engine) SiteInfo(c *web.Context) *web.HttpError {
	rst := make(map[string]interface{}, 0)
	for _, k := range []string{"title", "copyright", "keywords", "description"} {
		rst[k] = p.LocaleDao.Get(p.Db, c.Locale(), "site."+k)
	}
	return c.JSON(rst)
}

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

	messages := make(map[string]interface{}, 0)
	for _, item := range items {
		ss := strings.Split(item.Code, ".")
		sl := len(ss)
		tmp := messages
		for i, k := range ss {
			if i+1 == sl {
				tmp[k] = item.Message
			} else {
				if tmp[k] == nil {
					tmp[k] = make(map[string]interface{}, 0)
				}
				tmp = tmp[k].(map[string]interface{})
			}
		}
	}
	return c.JSON(messages)
}
