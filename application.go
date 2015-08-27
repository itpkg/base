package base

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"log/syslog"
	"os"
	"text/template"
	"time"

	"github.com/facebookgo/inject"
	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
)

type Application struct {
	beans inject.Graph
}

func (p *Application) Map(name string, value interface{}) {
	p.beans.Provide(&inject.Object{Value: value, Name: name})
}

func (p *Application) Load(file string, ping bool) error {
	logger, err := syslog.New(syslog.LOG_LOCAL7, "itpkg")
	if err != nil {
		return err
	}
	p.Map("logger", logger)

	_, err = os.Stat(file)
	if err != nil {
		return err
	}

	cfg := Configuration{}
	logger.Info(fmt.Sprintf("Load from config file: %s", file))

	var tmp *template.Template
	tmp, err = template.ParseFiles(file)
	if err != nil {
		return err
	}

	vars := make(map[string]interface{}, 0)

	vars["Env"] = os.Getenv("ITPKG_ENV")
	vars["Secrets"] = os.Getenv("ITPKG_SECRETS")
	vars["DbPassword"] = os.Getenv("ITPKG_DATABASE_PASSWORD")

	var buf bytes.Buffer

	if err = tmp.Execute(&buf, vars); err != nil {
		return err
	}
	if err = yaml.Unmarshal(buf.Bytes(), &cfg); err != nil {
		return err
	}
	if ping {
		if err = p.ping(&cfg); err != nil {
			return err
		}
	}
	p.Map("cfg", &cfg)
	return nil

}

func (p *Application) ping(cfg *Configuration) error {
	err := p.beans.Populate()
	if err != nil {
		return err
	}

	//database
	var db gorm.DB
	db, err = gorm.Open("postgres", cfg.DbUrl())
	if err != nil {
		return err
	}
	db.LogMode(!cfg.IsProduction())
	if err = db.DB().Ping(); err != nil {
		return err
	}
	db.DB().SetMaxIdleConns(12)
	db.DB().SetMaxOpenConns(120)
	p.Map("db", &db)

	//redis
	p.Map("redis", &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 4 * 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.RedisUrl())
			if err != nil {
				return nil, err
			}
			if _, err = c.Do("SELECT", cfg.Redis.Db); err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	})

	//helper
	var cip cipher.Block
	if cip, err = aes.NewCipher([]byte(cfg.Secrets[60:92])); err != nil {
		return err
	}
	p.Map("aes.cip", cip)
	p.Map("hmac.key", []byte(cfg.Secrets[20:84]))
	p.Map("helper", Helper{})

	err = p.beans.Populate()
	return err
}
