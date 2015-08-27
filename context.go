package base

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"log/syslog"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/facebookgo/inject"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jrallison/go-workers"
	"github.com/pborman/uuid"
	"gopkg.in/yaml.v2"
)

type Context struct {
}

func (p *Context) Load(file string, ping bool) error {
	logger, err := syslog.New(syslog.LOG_LOCAL7, "itpkg")
	if err != nil {
		return err
	}
	Map(logger)

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
	MapTo("base.cfg", &cfg)

	//helper
	var cip cipher.Block
	if cip, err = aes.NewCipher([]byte(cfg.Secrets[60:92])); err != nil {
		return err
	}
	MapTo("aes.cip", cip)
	MapTo("hmac.key", []byte(cfg.Secrets[20:84]))
	MapTo("base.helper", &Helper{})

	if ping {
		if err = p.ping(&cfg, logger); err != nil {
			return err
		}
	}

	err = beans.Populate()
	return err

}

func (p *Context) ping(cfg *Configuration, logger *syslog.Writer) error {

	//database
	db, err := gorm.Open("postgres", cfg.DbUrl())
	if err != nil {
		return err
	}
	db.LogMode(!cfg.IsProduction())
	if err = db.DB().Ping(); err != nil {
		return err
	}
	db.DB().SetMaxIdleConns(12)
	db.DB().SetMaxOpenConns(120)
	Map(&db)

	//redis
	Map(&redis.Pool{
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

	//workers
	workers.Configure(map[string]string{
		"server":   cfg.RedisUrl(),
		"database": strconv.Itoa(cfg.Redis.Db),
		"pool":     "12",
		"process":  uuid.New(),
	})
	workers.Middleware.Append(&JobMiddleware{logger: logger})

	//gin
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	Map(gin.Default())

	MapTo("base.app", &Application{})

	return nil
}

//-----------------------------------------------------------------------------

var beans inject.Graph

func Get(name string) interface{} {
	for _, o := range beans.Objects() {
		if name == o.Name {
			return o.Value
		}
	}
	return nil
}

func Map(value interface{}) {
	beans.Provide(&inject.Object{Value: value})
}

func MapTo(name string, value interface{}) {
	beans.Provide(&inject.Object{Value: value, Name: name})
}

func LoopEngine(fn func(en Engine) error) error {
	for _, o := range beans.Objects() {
		switch o.Value.(type) {
		case Engine:
			if e := fn(o.Value.(Engine)); e != nil {
				return e
			}
		default:

		}
	}
	return nil

}
