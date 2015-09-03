package base

import (
	"log/syslog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jrallison/go-workers"
	"github.com/pborman/uuid"
)

type AuthEngine struct {
	Helper  *Helper        `inject:"base.helper"`
	Db      *gorm.DB       `inject:""`
	Logger  *syslog.Writer `inject:""`
	Router  *gin.Engine    `inject:""`
	AuthDao *AuthDao       `inject:""`
}

func (p *AuthEngine) Cron() {

}

func (p *AuthEngine) Job() (string, func(message *workers.Msg), float32) {
	return "", nil, 0.0
}

func (p *AuthEngine) Mount() {
	MapTo("dao.auth", &AuthDao{})

	p.Router.GET("/users/sign_in", func(c *gin.Context) {
		locale := Locale(c)
		fm := NewForm("sign_in", "user", Url("/users/sign_in", locale))
		fm.AddEmailField("email", "")
		fm.AddPasswordField("password", false)
		c.JSON(http.StatusOK, fm)
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

func (p *AuthEngine) Info() (name string, version string, desc string) {
	return "auth", "v20150826", "auth module"
}

//-----------------------model---------------------------------------
type User struct {
	Model
	Uid       string `sql:"not null;type:UUID;default:uuid_generate_v4()"`
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
			ro != nil &&
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
	Role   *Role
	RoleID uint `sql:"index;not null"`
	DateZone
}

//-----------------------dao---------------------------------------

type AuthDao struct {
	Helper *Helper `inject:"base.helper"`
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
