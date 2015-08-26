package base

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"log"
	"time"

	"github.com/facebookgo/inject"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"
)

var beans inject.Graph

func Register(objects map[string]interface{}) error {
	items := make([]*inject.Object, 0)
	for k, v := range objects {
		items = append(items, &inject.Object{Value: v, Name: k})
	}
	return beans.Provide(items...)
}

func LoopEngine(f func(en Engine) error) error {
	for _, obj := range beans.Objects() {
		switch obj.Value.(type) {
		case Engine:
			// en := obj.Value.(Engine)
			// n,v,d := en.Info()
			// log.Printf("%s %s %s", n, v, d)
			if err := f(obj.Value.(Engine)); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

func New(file string) (*Application, error) {
	cfg, err := Load(file)
	if err != nil {
		return nil, err
	}

	if cfg.IsProduction() {
		if bkd, err := logging.NewSyslogBackend("itpkg"); err == nil {
			logging.SetBackend(bkd)
			logging.SetLevel(logging.INFO, "")
		} else {
			log.Fatalf("%v", err)
		}
	} else {
		logging.SetLevel(logging.DEBUG, "")
	}
	app := &Application{Cfg: cfg}

	args := make(map[string]interface{}, 0)

	//database
	var db gorm.DB
	db, err = gorm.Open("postgres", cfg.DbUrl())
	if err != nil {
		return nil, err
	}
	db.LogMode(!cfg.IsProduction())
	if err = db.DB().Ping(); err != nil {
		return nil, err
	}
	db.DB().SetMaxIdleConns(12)
	db.DB().SetMaxOpenConns(120)
	for _, ext := range []string{"uuid-ossp", "pgcrypto"} {
		db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\"", ext))
	}
	args["db"] = &db

	//redis
	args["redis"] = &redis.Pool{
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
	}

	//aes
	var cip cipher.Block
	if cip, err = aes.NewCipher([]byte(cfg.Secrets[60:92])); err != nil {
		log.Fatalf("error on generate aes cipher: %v", err)
	}
	args["aes"] = cip

	//logger
	args["logger"] = logging.MustGetLogger("itpkg")

	//router

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	args["router"] = gin.Default()

	//Init
	args["cfg"] = cfg
	args["app"] = app

	if err = Register(args); err != nil {
		return nil, err
	}
	if err = beans.Populate(); err != nil {
		return nil, err
	}

	return app, nil
}
