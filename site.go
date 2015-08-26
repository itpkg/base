package base

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type SiteEngine struct {
}

func (p *SiteEngine) Job() {

}
func (p *SiteEngine) Mount() {

}

func (p *SiteEngine) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Setting{})
	db.AutoMigrate(&Locale{})
	db.Model(&Locale{}).AddUniqueIndex("idx_locales_key_lang", "key", "lang")
}

func (p *SiteEngine) Seeds(*gorm.DB) {

}

func (p *SiteEngine) Info() (name string, version string, desc string) {
	return "site", "v20150826", "site framework"
}

//---------models
type Setting struct {
	ID  string `gorm:"primary_key"`
	Val []byte `sql:"not null"`
	Iv  []byte `sql:"size:32"`
}

type Locale struct {
	Key  string `sql:"not null;size:255;index"`
	Val  string `sql:"not null;type:TEXT"`
	Lang string `sql:"not null;size:5;index;default:'en'"`
}

//---------------daos
type SiteDao struct {
	Db  *gorm.DB
	Aes *Aes
}

func (p *SiteDao) Set(key string, val interface{}, enc bool) error {
	dt, err := Obj2bits(val)
	if err != nil {
		return err
	}
	var iv []byte
	if enc {
		dt, iv, err = p.Aes.Encrypt(dt)
		if err != nil {
			return err
		}
	}

	st := Setting{ID: key}
	var cn int
	p.Db.Model(st).Count(&cn)
	if cn == 0 {
		st.Val = dt
		st.Iv = iv
		p.Db.Create(&st)
	} else {
		p.Db.Model(&st).Updates(Setting{Val: dt, Iv: iv})
	}
	return nil
}

func (p *SiteDao) Get(key string, val interface{}, enc bool) error {
	st := Setting{}
	p.Db.Where("id = ?", key).First(&st)
	if st.Val != nil {
		var dt []byte

		if enc {
			dt = p.Aes.Decrypt(st.Val, st.Iv)
		} else {
			dt = st.Val
		}
		return Bits2obj(dt, val)
	}
	return nil
}

type LocaleDao struct {
	Db *gorm.DB
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
	p.Db.Where("lang = ? AND key = ?", lang, key).First(&l)
	return l.Val
}

func (p *LocaleDao) Set(lang, key, val string) {
	l := Locale{Lang: lang, Key: key}
	p.Db.Where("lang = ? AND key = ?", lang, key).First(&l)
	if l.Val == "" {
		p.Db.Create(&Locale{Key: key, Lang: lang, Val: val})
	} else {
		p.Db.Model(&l).Updates(Locale{Val: val})
	}

}
