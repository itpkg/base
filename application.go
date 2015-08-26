package base

import (
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"
)

type Application struct {
	Cfg    *Configuration  `inject:""`
	Logger *logging.Logger `inject:""`
	Router *gin.Engine     `inject:""`
	Redis  *redis.Pool     `inject:""`
	Db     *gorm.DB        `inject:""`
}

func (p *Application) LoopEngine(f func(en Engine) error) error {
	for _, obj := range beans.Objects() {
		//		if eo, ok := obj.Value.(Engine); ok {
		//			if err := f(eo); err != nil {
		//				return err
		//			}
		//			fmt.Sprintln("########")
		//		}

		switch obj.Value.(type) {
		case Engine:
			fmt.Printf("#### %s\n", obj)
			if err := f(obj.Value.(Engine)); err != nil {
				return err
			}
		default:
			//fmt.Printf("#### %s\n", obj.Value)
		}
	}
	return nil
}

func (p *Application) DbMigrate() error {
	return p.LoopEngine(func(en Engine) error {
		en.Migrate()
		return nil
	})
}

func (p *Application) Server() error {
	if err := p.LoopEngine(func(en Engine) error {
		en.Mount()
		return nil
	}); err != nil {
		return err
	}
	return http.ListenAndServe(fmt.Sprintf("%s:%d", p.Cfg.Http.Host, p.Cfg.Http.Port), p.Router)
}

func (p *Application) Openssl() error {
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
	return nil
}

func (p *Application) Nginx() error {
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
	return nil
}

func (p *Application) ClearRedis(pat string) error {

	r := p.Redis.Get()
	defer r.Close()

	v, e := r.Do("KEYS", pat)
	if e != nil {
		return e
	}
	ks := v.([]interface{})
	if len(ks) == 0 {
		p.Logger.Info("Empty!!!")
		return nil
	}
	_, e = r.Do("DEL", ks...)
	if e != nil {
		return e
	}
	p.Logger.Info("Clear redis keys by '%s' succressfully!", pat)
	return nil
}
