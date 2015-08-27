package base

import (
	"fmt"
	"log/syslog"
	"os"
	"reflect"
	"text/template"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
)

type Application interface {
	Server()
	Migrate()
	Seed() error
	Nginx()
	Openssl()
	Clear(pat string) error
	Register(en Engine)
}

type application struct {
	mrt     *martini.ClassicMartini
	engines []Engine
}

func (p *application) Server() {
	cfg := p.mrt.Injector.Get(reflect.TypeOf((*Configuration)(nil))).Interface().(*Configuration)
	p.mrt.RunOnAddr(fmt.Sprintf("%s:%d", cfg.Http.Host, cfg.Http.Port))
}

func (p *application) Migrate() {
	p.loop(func(en Engine) error {
		p.mrt.Invoke(en.Migrate)
		return nil
	})
}

func (p *application) Seed() error {
	return p.loop(func(en Engine) error {
		_, err := p.mrt.Invoke(en.Seed)
		return err
		//		if val, err := p.mrt.Invoke(en.Seed); err == nil {
		//
		//			ret := val[0].Interface()
		//			if ret == nil {
		//				return nil
		//			} else {
		//				return ret.(error)
		//			}
		//		} else {
		//			return err
		//		}

	})
}

func (p *application) loop(fn func(en Engine) error) error {
	for _, en := range p.engines {
		if err := fn(en); err != nil {
			return err
		}
	}
	return nil
}

func (p *application) Register(en Engine) {
	p.engines = append(p.engines, en)
}

func (p *application) Openssl() {
	args := make(map[string]interface{}, 0)

	//todo 加载domain
	args["domain"] = "localhost"
	t := template.Must(template.New("ssl.sh").Parse(
		`
openssl genrsa -out root-key.pem 2048
openssl req -new -key root-key.pem -out root-req.csr -text
openssl x509 -req -in root-req.csr -out root-cert.pem -sha512 -signkey root-key.pem -days 3650 -text -extfile /etc/ssl/openssl.cnf -extensions v3_ca

openssl genrsa -out {{.domain}}-key.pem 2048
openssl req -new -key {{.domain}}-key.pem -out {{.domain}}-req.csr -text
openssl x509 -req -in {{.domain}}-req.csr -CA root-cert.pem -CAkey root-key.pem -CAcreateserial -days 3650 -out {{.domain}}-cert.pem -text

openssl verify -CAfile root-cert.pem {{.domain}}-cert.pem
openssl rsa -noout -text -in {{.domain}}-key.pem
openssl req -noout -text -in {{.domain}}-req.csr
openssl x509 -noout -text -in {{.domain}}-cert.pem
`))

	t.Execute(os.Stdout, args)

}

func (p *application) Nginx() {
	cfg := p.mrt.Injector.Get(reflect.TypeOf((*Configuration)(nil))).Interface().(*Configuration)

	args := make(map[string]interface{}, 0)
	//todo 加载domain
	args["domain"] = "localhost"
	args["host"] = cfg.Http.Host
	args["port"] = cfg.Http.Port
	args["pwd"], _ = os.Getwd()

	t := template.Must(template.New("ssl.sh").Parse(

		`
upstream {{.domain}}.conf {
server http://{{.host}}:{{.port}} fail_timeout=0;
}


server {
listen 443;
ssl  on;
ssl_certificate  ssl/{{.domain}}-cert.pem;
ssl_certificate_key  ssl/{{.domain}}-key.pem;
ssl_session_timeout  5m;
ssl_protocols  SSLv2 SSLv3 TLSv1;
ssl_ciphers  RC4:HIGH:!aNULL:!MD5;
ssl_prefer_server_ciphers  on;

client_max_body_size 4G;
keepalive_timeout 10;

server_name {{.domain}} www.{{.domain}};

root {{.pwd}}/public;
try_files $uri $uri/index.html @{{.domain}}.conf;

location @{{.domain}}.conf {
proxy_set_header X-Forwarded-Proto https;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header Host $http_host;
proxy_set_header  X-Real-IP $remote_addr;
proxy_redirect off;
proxy_pass http://{{.domain}}.conf;
# limit_req zone=one;
access_log log/{{.domain}}.access.log;
error_log log/{{.domain}}.error.log;
}

location ~* \.(?:css|js|html|jpg|jpeg|gif|png|ico)$ {
gzip_static on;
expires max;
add_header Cache-Control public;
}

location = /50x.html {
root html;
}

location = /404.html {
root html;
}

location @503 {
error_page 405 = /system/maintenance.html;
if (-f $document_root/system/maintenance.html) {
rewrite ^(.*)$ /system/maintenance.html break;
}
rewrite ^(.*)$ /503.html break;
}

if ($request_method !~ ^(GET|HEAD|PUT|PATCH|POST|DELETE|OPTIONS)$ ){
return 405;
}

if (-f $document_root/system/maintenance.html) {
return 503;
}

location ~ \.(php|jsp|asp)$ {
return 405;
}

}
`))

	t.Execute(os.Stdout, args)

}

func (p *application) Clear(pat string) error {
	redis := p.mrt.Injector.Get(reflect.TypeOf((*redis.Pool)(nil))).Interface().(*redis.Pool)
	r := redis.Get()
	defer r.Close()

	v, e := r.Do("KEYS", pat)
	if e != nil {
		return e
	}
	ks := v.([]interface{})
	if len(ks) == 0 {
		return nil
	}
	_, e = r.Do("DEL", ks...)
	if e != nil {
		return e
	}
	return nil
}

//-----------------------------------------------------------------------------
func New(name string) (Application, error) {
	mrt := martini.Classic()

	//configuration
	cfg, err := Load(name)
	if err != nil {
		return nil, err
	}
	mrt.Map(cfg)

	//logger
	var logger *syslog.Writer
	if logger, err = syslog.New(syslog.LOG_LOCAL7, "itpkg"); err != nil {
		return nil, err
	}
	mrt.Map(logger)
	mrt.Use(Logger())

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
	mrt.Map(&db)

	//redis
	mrt.Map(&redis.Pool{
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

	//aes
	aes := Aes{}
	if err = aes.Init([]byte(cfg.Secrets[60:92])); err != nil {
		return nil, err
	}
	mrt.Map(&aes)

	//hmac
	hmac := Hmac{}
	hmac.Init([]byte(cfg.Secrets[20:52]))
	mrt.Map(&hmac)

	app := &application{
		mrt:     mrt,
		engines: make([]Engine, 0),
	}

	app.Register(&SiteEngine{})
	app.Register(&AuthEngine{})

	app.loop(func(en Engine) error {
		en.Mount(mrt)
		return nil
	})
	return app, nil

}
