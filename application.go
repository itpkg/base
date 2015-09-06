package base

import (
	"fmt"
	"log/syslog"
	"os"
	"text/template"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jrallison/go-workers"
)

type Application struct {
	Db     *gorm.DB       `inject:""`
	Redis  *redis.Pool    `inject:""`
	Logger *syslog.Writer `inject:""`
	Cfg    *Configuration `inject:"base.cfg"`
	Router *gin.Engine    `inject:""`
	Helper *Helper        `inject:"base.helper"`
}

func (p *Application) Dispatcher() {
	LoopEngine(func(en Engine) error {
		en.Cron()
		return nil
	})

}

func (p *Application) Worker(port, threads int) {

	p.Logger.Info("Startup worker progress")

	LoopEngine(func(en Engine) error {
		queue, call, pri := en.Job()
		if queue != "" {
			workers.Process(queue, call, int(float32(threads)*pri)+1)
		}
		return nil
	})

	p.Logger.Info(fmt.Sprintf("Stats will be available at http://localhost:%d/stats", port))
	go workers.StatsServer(port)

	workers.Run()
}

func (p *Application) Server() {
	ro := p.Router
	ro.Use(SetLocale())
	ro.Use(SetTransactions(p.Db))
	ro.Use(SetCurrentUser(p.Helper, p.Cfg, p.Logger))

	LoopEngine(func(en Engine) error {
		en.Mount()
		return nil
	})

	ro.Run(p.Cfg.HttpUrl())
}

func (p *Application) Routes() {
	//todo
	for _, h := range p.Router.Handlers {
		fmt.Printf("%v", h)
	}
}

func (p *Application) Migrate() {
	for _, ext := range []string{"pgcrypto"} {
		p.Db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\"", ext))
	}

	LoopEngine(func(en Engine) error {
		en.Migrate()
		return nil
	})
}

func (p *Application) Seed() error {
	return LoopEngine(func(en Engine) error {
		return en.Seed()

	})
}

func (p *Application) Openssl() {
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

func (p *Application) Nginx() {

	args := make(map[string]interface{}, 0)
	//todo 加载domain
	args["domain"] = "localhost"
	args["host"] = p.Cfg.Http.Host
	args["port"] = p.Cfg.Http.Port
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

func (p *Application) Clear(pat string) error {
	r := p.Redis.Get()
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
