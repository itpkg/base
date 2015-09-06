package base

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jrallison/go-workers"
	"github.com/op/go-logging"
	"github.com/pborman/uuid"
)

type AuthEngine struct {
	Helper  *Helper         `inject:"base.helper"`
	Cfg     *Configuration  `inject:"base.cfg"`
	Db      *gorm.DB        `inject:""`
	Logger  *logging.Logger `inject:""`
	Router  *gin.Engine     `inject:""`
	AuthDao *AuthDao        `inject:""`
	I18n    *I18n           `inject:""`
}

func (p *AuthEngine) Cron() {

}

func (p *AuthEngine) Job() (string, func(message *workers.Msg), float32) {
	return "", nil, 0.0
}

func (p *AuthEngine) Nav(admin bool) []*Link {
	root, _, _ := p.Info()
	links := []*Link{
		&Link{Url: "/" + root + "/profile", Label: "form.title.user.profile"},
		&Link{Url: "/" + root + "/logs", Label: "link.title.personal.logs"},
	}

	return links
}

func (p *AuthEngine) Mount() {
	MapTo("dao.auth", &AuthDao{})
	root, _, _ := p.Info()

	rt := p.Router.Group("/" + root)

	rt.GET("/self", func(c *gin.Context) {
		user := CurrentUser(c)
		nb := NewNavBar()

		if user != nil {
			admin := user.Is("admin")
			LoopEngine(func(en Engine) error {
				if nav := en.Nav(admin); nav != nil {
					path, _, _ := en.Info()
					dd := NewDropDown("engine." + path + ".name")
					dd.AddLinks(nav)

					nb.Add(dd)
				}
				return nil

			})
		}

		RESPONSE(c, p.I18n, nb)
	})

	rt.GET("/bar", func(c *gin.Context) {
		links := NewDropDown("")
		locale := Locale(c)
		user := CurrentUser(c)
		if user == nil {
			links.Label = "label.sign_in_or_up"
			for _, v := range []string{"sign_in", "sign_up", "forgot_password", "confirm", "unlock"} {
				links.Add("/"+root+"/"+v, "form.title.user."+v)
			}

		} else {
			links.Label = "label.welcome"
			links.Add("/"+root+"/self", "link.title.settings")
			links.Add("/"+root+"/sign_out", "form.title.user.sign_out")

		}
		links.T(p.I18n, locale)
		if user != nil {
			links.Label += user.Name
		}
		c.JSON(http.StatusOK, links)
	})

	rt.GET("/sign_in", func(c *gin.Context) {
		fm := NewForm("user.sign_in", "/"+root+"/sign_in")
		fm.AddEmailField("email", "", true, false)
		fm.AddPasswordField("password", true, false)
		RESPONSE(c, p.I18n, fm)
	})

	rt.POST("/sign_in", func(c *gin.Context) {
		locale := Locale(c)
		db := Db(c)
		var fm SignInFm
		res := NewResponse()

		if err := c.Bind(&fm); err == nil {
			user := p.AuthDao.Auth(db, fm.Email, fm.Password)
			if user == nil {
				res.AddError("error.user.email_password_not_match")
			} else {
				tkm := make(map[string]interface{}, 0)
				tkm["user"] = user.Uid
				p.AuthDao.Log(db, user.ID, p.I18n.T(locale, "log.user.sign_in"), "info")
				if tk, err := p.Helper.TokenCreate(tkm, p.Cfg.Http.Expire); err == nil {
					res.AddData(tk)
				} else {
					res.AddError(err.Error())
				}
			}

		} else {
			res.AddError("error.inputs_invalid")
		}
		RESPONSE(c, p.I18n, res)
	})

	rt.GET("/sign_up", func(c *gin.Context) {
		fm := NewForm("user.sign_up", "/"+root+"/sign_up")
		fm.AddTextField("username", "", true, false)
		fm.AddEmailField("email", "", true, false)
		fm.AddPasswordField("password", true, true)
		RESPONSE(c, p.I18n, fm)
	})
	rt.POST("/sign_up", func(c *gin.Context) {
		//todo
	})

	rt.GET("/confirm", func(c *gin.Context) {
		fm := NewForm("user.confirm", "/"+root+"/sign_up")
		fm.AddEmailField("email", "", true, false)
		RESPONSE(c, p.I18n, fm)
	})
	rt.POST("/confirm", func(c *gin.Context) {
		//todo
	})

	rt.GET("/unlock", func(c *gin.Context) {
		fm := NewForm("user.unlock", "/"+root+"/unlock")
		fm.AddEmailField("email", "", true, false)
		RESPONSE(c, p.I18n, fm)
	})
	rt.POST("/unlock", func(c *gin.Context) {
		//todo
	})

	rt.GET("/forgot_password", func(c *gin.Context) {
		fm := NewForm("user.forgot_password", "/"+root+"/forgot_password")
		fm.AddEmailField("email", "", true, false)
		RESPONSE(c, p.I18n, fm)
	})
	rt.POST("/forgot_password", func(c *gin.Context) {
		//todo
	})

	rt.GET("/change_password", func(c *gin.Context) {
		fm := NewForm("user.forgot_password", "/"+root+"/change_password")
		fm.AddHiddenField("token", c.Query("token"))
		fm.AddPasswordField("password", true, true)
		RESPONSE(c, p.I18n, fm)
	})
	rt.POST("/change_password", func(c *gin.Context) {
		//todo
	})

	rt.GET("/profile", func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{})
		} else {
			fm := NewForm("user.profile", "/"+root+"/profile")
			fm.AddTextField("username", user.Name, true, false)
			fm.AddPasswordField("current_password", true, false)
			fm.AddPasswordField("new_password", true, true)
			RESPONSE(c, p.I18n, fm)
		}
	})
	rt.POST("/profile", func(c *gin.Context) {
		locale := Locale(c)
		db := Db(c)
		var fm ProfileFm
		res := NewResponse()
		if err := c.Bind(&fm); err == nil &&
			fm.Password == fm.ConfirmPassword &&
			(fm.Password == "" || len(fm.Password) >= 6) {

			user := CurrentUser(c)
			if user != nil && p.Helper.HmacEqual([]byte(fm.CurrentPassword), user.Password) {

				db.Model(user).Updates(User{Name: fm.Username})
				if fm.Password != "" {
					db.Model(user).Updates(User{Password: p.Helper.HmacSum([]byte(fm.Password))})
				}

				p.AuthDao.Log(db, user.ID, p.I18n.T(locale, "log.user.update_profile"), "info")

			} else {
				res.AddError("error.user.bad_password")
			}

		} else {
			res.AddError("error.inputs_invalid")
		}
		RESPONSE(c, p.I18n, res)
	})

	rt.GET("/sign_out", func(c *gin.Context) {
		locale := Locale(c)
		db := Db(c)
		user := CurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{})
		} else {
			p.Helper.TokenTtl(c.Request, 0)
			p.AuthDao.Log(db, user.ID, p.I18n.T(locale, "log.user.sign_out"), "info")
			c.JSON(http.StatusOK, gin.H{})
		}
	})
}

func (p *AuthEngine) Migrate() {
	db := p.Db
	db.AutoMigrate(&Contact{})
	db.AutoMigrate(&User{})
	db.Model(&User{}).AddUniqueIndex("idx_users_login", "token", "provider")
	db.AutoMigrate(&Log{})
	db.AutoMigrate(&Role{})
	db.Model(&Role{}).AddUniqueIndex("idx_roles_name_resource", "name", "resource_type", "resource_id")
	db.AutoMigrate(&Permission{})
	db.Model(&Permission{}).AddUniqueIndex("idx_permissions_role_user", "role_id", "user_id")
}

func (p *AuthEngine) Seed() error {

	tx := p.Db.Begin()
	email := "root@localhost"
	if p.AuthDao.GetByEmail(tx, email) == nil {
		var user *User
		user = p.AuthDao.CreateByEmail(tx, email, "root", "changeme")

		p.AuthDao.Confirm(tx, user.ID)
		p.AuthDao.AddRole(tx, user.ID, "root", "", 0, nil, nil)
		p.AuthDao.AddRole(tx, user.ID, "admin", "", 0, nil, nil)

	}
	tx.Commit()
	return nil
}

func (p *AuthEngine) Info() (name, version, desc string) {
	return "personal", "v20150826", "auth module"
}

//-----------------------form---------------------------------------
type SignInFm struct {
	Email    string `form:"email" binding:"required"`
	Password string `form:"password" binding:"required"`
}
type ChangePasswordFm struct {
	Token      string `form:"token"`
	Password   string `form:"password"`
	RePassword string `form:"re_password"`
}
type SignUpFm struct {
	Username        string `form:"username" binding:"required"`
	Email           string `form:"email" binding:"required"`
	CurrentPassword string `form:"current_password"`
	Password        string `form:"password"`
	RePassword      string `form:"re_password"`
}
type ProfileFm struct {
	Username        string `form:"username" binding:"required"`
	CurrentPassword string `form:"current_password" json:"current_password" binding:"required"`
	Password        string `form:"new_password" json:"new_password" binding:""`
	ConfirmPassword string `form:"re_new_password" json:"re_new_password" binding:""`
}
type EmailFm struct {
	Email string `form:"email" binding:"required"`
}

//-----------------------model---------------------------------------
type User struct {
	Model
	Uid       string `sql:"not null;type:UUID;default:gen_random_uuid()"`
	Name      string `sql:"not null;size:64;index"`
	Email     string `sql:"size:128;index"`
	Token     string `sql:"size:255;index;not null"`
	Provider  string `sql:"size:16;not null;default:'email';index"`
	Password  []byte `sql:"size:64"`
	Confirmed *time.Time
	Locked    *time.Time

	ContactID   uint `sql:"not null"`
	Contact     Contact
	Logs        []Log
	Permissions []Permission
}

func (p *User) Is(name string) bool {
	return p.Has(name, "", 0)
}

func (p User) Can(name, resource_type string) bool {
	return p.Has(name, resource_type, 0)
}

func (p *User) Has(name, resource_type string, resource_id uint) bool {
	now := time.Now()
	for _, pm := range p.Permissions {
		ro := pm.Role
		if pm.StartUp.Before(now) &&
			pm.ShutDown.After(now) &&
			ro.Name == name &&
			ro.ResourceType == resource_type &&
			ro.ResourceID == resource_id {
			return true
		}
	}
	return false
}

type Contact struct {
	Model
	Qq       string
	Skype    string
	WeChat   string
	LinkedIn string
	FaceBook string
	Email    string
	Logo     string
	Phone    string
	Tel      string
	Fax      string
	Address  string
	Details  string `sql:"type:TEXT"`
}

type Log struct {
	ID        uint
	UserID    uint   `sql:"not null;index"`
	Message   string `sql:"size:255"`
	Type      string `sql:"size:8;default:'info';index"`
	CreatedAt time.Time
}

type Role struct {
	ID           uint
	Name         string `sql:"size:255;index;not null"`
	ResourceType string `sql:"size:255;index;not null"`
	ResourceID   uint   `sql:"index;not null"`
	Permissions  []Permission
}

type Permission struct {
	Model
	User   User
	UserID uint `sql:"index;not null"`
	Role   Role
	RoleID uint `sql:"index;not null"`
	DateZone
}

//-----------------------dao---------------------------------------

type AuthDao struct {
	Helper *Helper         `inject:"base.helper"`
	Logger *logging.Logger `inject:""`
}

func (p *AuthDao) Auth(db *gorm.DB, email, password string) *User {
	if user := p.GetByEmail(db, email); user != nil && p.Helper.HmacEqual([]byte(password), user.Password) {
		return user
	}
	return nil
}

func (p *AuthDao) GetByEmail(db *gorm.DB, email string) *User {
	var user User
	if db.Model(User{}).Where("email = ? AND provider = ?", email, "email").First(&user).RecordNotFound() {
		return nil
	}
	return &user
}

func (p *AuthDao) CreateByEmail(db *gorm.DB, email, name, password string) *User {
	u := User{
		Provider: "email",
		Name:     name,
		Email:    email,
		Password: p.Helper.HmacSum([]byte(password)),
		Token:    email,
		Contact:  Contact{},
	}
	db.Create(&u)
	return &u
}

func (p *AuthDao) ResetUid(db *gorm.DB, user uint) {
	db.Model(User{}).Where("id = ?", user).Updates(User{Uid: uuid.New()})
}

func (p *AuthDao) Confirm(db *gorm.DB, user uint) {
	now := time.Now()
	db.Model(User{}).Where("id = ?", user).Updates(User{Confirmed: &now})
}

func (p *AuthDao) Log(db *gorm.DB, user uint, message string, flag string) {
	db.Create(&Log{UserID: user, Message: message, Type: flag})
}

func (p *AuthDao) AddRole(db *gorm.DB, user uint, name string, resource_type string, resource_id uint, startUp, shutDown *time.Time) {
	role := p.getRole(db, name, resource_type, resource_id)
	for _, pe := range role.Permissions {
		if pe.UserID == user {
			return
		}
	}
	db.Create(&Permission{
		UserID: user,
		RoleID: role.ID,
		DateZone: DateZone{
			StartUp:  startUp,
			ShutDown: shutDown,
		},
	})
}

func (p *AuthDao) DelRole(db *gorm.DB, user uint, name string, resource_type string, resource_id uint) {
	role := p.getRole(db, name, resource_type, resource_id)
	for _, pe := range role.Permissions {
		if pe.UserID == user {
			db.Delete(&pe)
			return
		}
	}

}
func (p *AuthDao) Can(db *gorm.DB, user uint, name string, resource_type string, resource_id uint) bool {
	role := p.getRole(db, name, resource_type, resource_id)
	for _, pe := range role.Permissions {
		if pe.UserID == user {
			return true
		}
	}
	return false
}

func (p *AuthDao) getRole(db *gorm.DB, name string, resource_type string, resource_id uint) *Role {
	role := Role{}
	db.FirstOrCreate(&role, Role{Name: name, ResourceType: resource_type, ResourceID: resource_id})

	db.Model(&role).Related(&role.Permissions)
	return &role
}

func init() {
	Map(&AuthEngine{})
}
