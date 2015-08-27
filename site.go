package base

import (
	"fmt"
	"io/ioutil"

	"github.com/jinzhu/gorm"
	"github.com/magiconair/properties"
)

type SiteEngine struct {
}

func (p *SiteEngine) Job() {

}
func (p *SiteEngine) Mount() {

}

func (p *SiteEngine) Migrate(db *gorm.DB) {
	for _, ext := range []string{"uuid-ossp", "pgcrypto"} {
		db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\"", ext))
	}

	db.AutoMigrate(&Setting{})
	db.AutoMigrate(&Locale{})
	db.Model(&Locale{}).AddUniqueIndex("idx_locales_key_lang", "key", "lang")
}

func (p *SiteEngine) Seed(db *gorm.DB, aes *Aes, hmac *Hmac) error {
	tx := db.Begin()
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
	db  *gorm.DB
	aes *Aes
}

func (p *SiteDao) Set(key string, val interface{}, enc bool) error {
	dt, err := Obj2bits(val)
	if err != nil {
		return err
	}
	var iv []byte
	if enc {
		dt, iv, err = p.aes.Encrypt(dt)
		if err != nil {
			return err
		}
	}

	st := Setting{ID: key}
	var cn int
	p.db.Model(st).Count(&cn)
	if cn == 0 {
		st.Val = dt
		st.Iv = iv
		p.db.Create(&st)
	} else {
		p.db.Model(&st).Updates(Setting{Val: dt, Iv: iv})
	}
	return nil
}

func (p *SiteDao) Get(key string, val interface{}, enc bool) error {
	st := Setting{}
	p.db.Where("id = ?", key).First(&st)
	if st.Val != nil {
		var dt []byte

		if enc {
			dt = p.aes.Decrypt(st.Val, st.Iv)
		} else {
			dt = st.Val
		}
		return Bits2obj(dt, val)
	}
	return nil
}

type LocaleDao struct {
	db *gorm.DB
}

func (p *LocaleDao) T(lang, key string, args ...interface{}) string {
	v := p.Get(lang, key)
	if v == "" {
		return fmt.Sprintf("Translation [%s] not found", key)
	}
	return fmt.Sprintf(v, args...)
}

func (p *LocaleDao) Get(lang, key string) string {
	l := Locale{Lang: lang, Key: key}
	p.db.Where("lang = ? AND key = ?", lang, key).First(&l)
	return l.Val
}

func (p *LocaleDao) Set(lang, key, val string) {
	l := Locale{Lang: lang, Key: key}
	p.db.Where("lang = ? AND key = ?", lang, key).First(&l)
	if l.Val == "" {
		p.db.Create(&Locale{Key: key, Lang: lang, Val: val})
	} else {
		p.db.Model(&l).Updates(Locale{Val: val})
	}

}
