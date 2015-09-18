package base

import (
	"github.com/garyburd/redigo/redis"
	"github.com/itpkg/web"
	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"
)

type Engine struct {
	Mux        *web.Mux        `inject:""`
	Logger     *logging.Logger `inject:""`
	Redis      *redis.Pool     `inject:""`
	Db         *gorm.DB        `inject:""`
	AuthDao    *AuthDao        `inject:""`
	LocaleDao  *LocaleDao      `inject:""`
	SettingDao *SettingDao     `inject:""`
}

func (p *Engine) Migrate() error {
	db := p.Db

	db.AutoMigrate(&Setting{})
	db.AutoMigrate(&Locale{})
	db.Model(&Locale{}).AddUniqueIndex("idx_locales_key", "lang", "code")

	db.AutoMigrate(&Contact{})
	db.AutoMigrate(&User{})
	db.Model(&User{}).AddUniqueIndex("idx_users_login", "token", "provider")
	db.AutoMigrate(&Log{})
	db.AutoMigrate(&Role{})
	db.Model(&Role{}).AddUniqueIndex("idx_roles_name_resource", "name", "resource_type", "resource_id")
	db.AutoMigrate(&Permission{})
	db.Model(&Permission{}).AddUniqueIndex("idx_permissions_role_user", "role_id", "user_id")

	return nil
}

func (p *Engine) Seed() error {
	tx := p.Db.Begin()
	email := "root@localhost"
	if p.AuthDao.GetByEmail(tx, email) == nil {
		var user *User
		user = p.AuthDao.CreateByEmail(tx, email, "root", "changeme")

		p.AuthDao.Confirm(tx, user.ID)
		p.AuthDao.AddRole(tx, user.ID, "root", "", 0, nil, nil)
		p.AuthDao.AddRole(tx, user.ID, "admin", "", 0, nil, nil)

	}
	p.LocaleDao.Load(tx, "locales")
	tx.Commit()

	return nil
}

func (p *Engine) Mount() {
	router := web.NewRouter()
	router.GET("^/sitemap.xml.gz$", p.Sitemap)
	router.GET("^/rss.atom$", p.Rss)
	router.GET(`^/locales/(?P<locale>[a-zA-Z_]{5})/translation.json$`, p.Locales)

	p.Mux.AddRouter(router)
}

func (p Engine) Info() (string, string) {
	return "base", "v20150917"
}

//-----------------------------------------------------------------------------
func init() {
	web.Register(&Engine{})
}
