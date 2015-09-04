package base

import (
	"fmt"
	"io/ioutil"
	"log/syslog"

	"github.com/garyburd/redigo/redis"
	"github.com/magiconair/properties"
)

type I18n struct {
	Redis  *redis.Pool    `inject:""`
	Logger *syslog.Writer `inject:""`
}

func (p *I18n) Load(path string) error {
	if files, err := ioutil.ReadDir(path); err == nil {
		for _, f := range files {
			fn := f.Name()
			lang := fn[0:(len(fn) - 11)]
			prop := properties.MustLoadFile(path+"/"+fn, properties.UTF8)
			for _, k := range prop.Keys() {
				p.Logger.Info(fmt.Sprintf("Find %s.%s]", lang, k))
				if err = p.Set(lang, k, prop.MustGetString(k)); err != nil {
					return err
				}
			}
		}
		return nil
	} else {
		return err
	}
}

func (p *I18n) T(lang, key string, args ...interface{}) string {
	if val, err := p.Get(lang, key); err == nil {
		return fmt.Sprintf(val, args...)
	} else {
		msg := fmt.Sprintf("Translation [%s] not found", key)
		p.Logger.Err(msg)
		return msg
	}
}

func (p *I18n) Get(lang, key string) (string, error) {
	c := p.Redis.Get()
	defer c.Close()
	if val, err := redis.String(c.Do("GET", p.id(lang, key))); err == nil {
		return val, nil
	} else {
		return "", err
	}
}

func (p *I18n) Set(lang, key, val string) error {
	c := p.Redis.Get()
	defer c.Close()
	_, err := c.Do("SET", p.id(lang, key), val)
	return err
}

func (p *I18n) id(lang, key string) string {
	return fmt.Sprintf("locales://%s/%s", lang, key)
}
