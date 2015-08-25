package base

import (
	"bytes"
	"os"
	"text/template"
	//	"fmt"
	//	"strconv"
	//	"time"
	//
	//	"github.com/garyburd/redigo/redis"
	//	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
)

type envVars struct {
	Env        string
	Secrets    string
	DbPassword string
}

type Configuration struct {
	Env     string
	Secrets string
	Http    struct {
		Host string
		Port int
	}
	Database struct {
		Adapter  string
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		Extra    string
	}
	Redis struct {
		Host string
		Port int
		Db   int
		Pool int
	}
}

//-----------------------------------------------------------------------------
func Load(file string) (*Configuration, error) {
	_, err := os.Stat(file)
	if err == nil {
		config := Configuration{}
		log.Info("Load from config file: %s", file)

		var tmp *template.Template
		tmp, err = template.ParseFiles(file)
		if err != nil {
			return nil, err
		}

		vars := envVars{
			Env:        os.Getenv("ITPKG_ENV"),
			Secrets:    os.Getenv("ITPKG_SECRETS"),
			DbPassword: os.Getenv("ITPKG_DATABASE_PASSWORD"),
		}
		var buf bytes.Buffer

		if err = tmp.Execute(&buf, vars); err != nil {
			return nil, err
		}
		if err = yaml.Unmarshal(buf.Bytes(), &config); err != nil {
			return nil, err
		}
		return &config, nil
	}
	return nil, err
}
