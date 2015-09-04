package base

import (
	"fmt"
	"log/syslog"
	"net/http"

	"github.com/carlescere/scheduler"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jrallison/go-workers"
)

type SiteEngine struct {
	Db         *gorm.DB       `inject:""`
	Logger     *syslog.Writer `inject:""`
	Redis      *redis.Pool    `inject:""`
	Router     *gin.Engine    `inject:""`
	I18n       *I18n          `inject:""`
	SettingDao *SettingDao    `inject:""`
}

func (p *SiteEngine) Cron() {
	job := func() {
		p.Logger.Info("Generate sitemap.xml")
		//todo
		p.Logger.Info("Generate rss.atom")
		//todo
	}
	scheduler.Every().Day().At("03:00").Run(job)
}

func (p *SiteEngine) Job() (string, func(message *workers.Msg), float32) {
	return "email", func(message *workers.Msg) {
		//todo
	}, 0.1
}

func (p *SiteEngine) Mount() {
	p.Router.GET("/info", func(c *gin.Context) {
		locale := Locale(c)
		db := Db(c)

		vals := make(map[string]interface{}, 0)
		for _, v := range c.Request.URL.Query()["keys"] {
			switch {
			case v == "title" || v == "keywords" || v == "description" || v == "author" || v == "copyright":
				vals[v] = p.I18n.T(locale, fmt.Sprintf("site.%s", v))
			case v == "links":
				links := make([]Link, 0)
				p.SettingDao.Get(db, "site.links", links, false)
				vals[v] = links
			}
		}
		c.JSON(http.StatusOK, vals)
	})
}

func (p *SiteEngine) Migrate() {
	db := p.Db
	db.AutoMigrate(&Setting{})
}

func (p *SiteEngine) Seed() error {
	return p.I18n.Load("locales")
}

func (p *SiteEngine) Info() (name, version, desc string) {
	return "site", "v20150826", "site module"
}

func (p *SiteEngine) Nav(admin bool) []*Link {
	if !admin {
		return nil
	}
	root, _, _ := p.Info()
	links := make([]*Link, 0)
	for _, v := range []string{"info", "seo", "captcha", "locales", "status"} {
		links = append(links, &Link{Url: "/" + root + "/" + v, Label: "link.title.site." + v})
	}
	return links
}

//---------models
type Setting struct {
	ID  string `gorm:"primary_key"`
	Val []byte `sql:"not null"`
	Iv  []byte `sql:"size:32"`
}

//---------------daos
type SettingDao struct {
	Helper *Helper `inject:"base.helper"`
}

func (p *SettingDao) Set(db *gorm.DB, key string, val interface{}, enc bool) error {
	dt, err := p.Helper.Obj2bits(val)
	if err != nil {
		db.Rollback()
		return err
	}
	var iv []byte
	if enc {
		dt, iv, err = p.Helper.AesEncrypt(dt)
		if err != nil {
			db.Rollback()
			return err
		}
	}

	st := Setting{ID: key}
	var cn int
	db.Model(st).Count(&cn)
	if cn == 0 {
		st.Val = dt
		st.Iv = iv
		db.Create(&st)
	} else {
		db.Model(&st).Updates(Setting{Val: dt, Iv: iv})
	}
	return nil
}

func (p *SettingDao) Get(db *gorm.DB, key string, val interface{}, enc bool) error {
	st := Setting{}
	db.Where("id = ?", key).First(&st)
	if st.Val != nil {
		var dt []byte

		if enc {
			dt = p.Helper.AesDecrypt(st.Val, st.Iv)
		} else {
			dt = st.Val
		}
		return p.Helper.Bits2obj(dt, val)
	}
	return nil
}

func init() {
	Map(&SiteEngine{})
}
