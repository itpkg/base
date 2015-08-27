package base

import (
	"fmt"
	"io/ioutil"
	"log/syslog"
	"net/http"

	"github.com/carlescere/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jrallison/go-workers"
	"github.com/magiconair/properties"
)

type SiteEngine struct {
	Db     *gorm.DB       `inject:""`
	Logger *syslog.Writer `inject:""`
	Router *gin.Engine    `inject:""`
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
	p.Router.GET("/site/:key", func(c *gin.Context) {
		//todo
		c.String(http.StatusOK, "Hello "+c.Param("key"))
	})
}

func (p *SiteEngine) Migrate() {
	db := p.Db

	db.AutoMigrate(&Setting{})
	db.AutoMigrate(&Locale{})
	db.Model(&Locale{}).AddUniqueIndex("idx_locales_key_lang", "key", "lang")
}

func (p *SiteEngine) Seed() error {
	tx := p.Db.Begin()
	path := "locales"
	if files, err := ioutil.ReadDir(path); err == nil {
		for _, f := range files {
			fn := f.Name()
			lang := fn[0:(len(fn) - 11)]
			prop := properties.MustLoadFile(path+"/"+fn, properties.UTF8)
			for _, k := range prop.Keys() {
				if err = tx.Create(&Locale{Lang: lang, Key: k, Val: prop.MustGetString(k)}).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
		tx.Commit()
		return nil
	} else {
		return err
	}
}

func (p *SiteEngine) Info() (name string, version string, desc string) {
	return "site", "v20150826", "site module"
}

//---------models
type Setting struct {
	ID  string `gorm:"primary_key"`
	Val []byte `sql:"not null"`
	Iv  []byte `sql:"size:32"`
}

type Locale struct {
	ID   uint   `gorm:"primary_key"`
	Key  string `sql:"not null;size:255;index"`
	Val  string `sql:"not null;type:TEXT"`
	Lang string `sql:"not null;size:5;index;default:'en_US'"`
}

//---------------daos
type SiteDao struct {
	Helper *Helper `inject:"base.helper"`
}

func (p *SiteDao) Set(db *gorm.DB, key string, val interface{}, enc bool) error {
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

func (p *SiteDao) Get(db *gorm.DB, key string, val interface{}, enc bool) error {
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

type LocaleDao struct {
}

func (p *LocaleDao) T(db *gorm.DB, lang, key string, args ...interface{}) string {
	v := p.Get(db, lang, key)
	if v == "" {
		return fmt.Sprintf("Translation [%s] not found", key)
	}
	return fmt.Sprintf(v, args...)
}

func (p *LocaleDao) Get(db *gorm.DB, lang, key string) string {
	l := Locale{Lang: lang, Key: key}
	db.Where("lang = ? AND key = ?", lang, key).First(&l)
	return l.Val
}

func (p *LocaleDao) Set(db *gorm.DB, lang, key, val string) {
	l := Locale{Lang: lang, Key: key}
	db.Where("lang = ? AND key = ?", lang, key).First(&l)
	if l.Val == "" {
		db.Create(&Locale{Key: key, Lang: lang, Val: val})
	} else {
		db.Model(&l).Updates(Locale{Val: val})
	}

}

func init() {
	Map(&SiteEngine{})
}
