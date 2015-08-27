package base

import (
	"errors"
	"reflect"
	"time"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/pborman/uuid"
)

type AuthEngine struct {
	db   *gorm.DB
	hmac *Hmac
}

func (p *AuthEngine) Job() {

}
func (p *AuthEngine) Mount(mrt *martini.ClassicMartini) {
	p.db = mrt.Injector.Get(reflect.TypeOf((*gorm.DB)(nil))).Interface().(*gorm.DB)
	p.hmac = mrt.Injector.Get(reflect.TypeOf((*Hmac)(nil))).Interface().(*Hmac)

}
func (p *AuthEngine) Migrate() {
	db := p.db
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
	tx := p.db.Begin()
	authDao := AuthDao{db: tx, hmac: p.hmac}

	user, err := authDao.AddByEmail("root@localhost", "root", "changeme")
	if err == nil {
		authDao.Confirm(user.ID)
		authDao.AddRole(user.ID, "root", "", 0, nil, nil)
		authDao.AddRole(user.ID, "admin", "", 0, nil, nil)
	}
	tx.Commit()
	return err
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
	db   *gorm.DB
	hmac *Hmac
}

func (p *AuthDao) AddByEmail(email, name, password string) (*User, error) {
	var c int
	p.db.Model(User{}).Where("email = ? AND provider = ?", email, "email").Count(&c)
	if c > 0 {
		p.db.Rollback()
		return nil, errors.New("email already exist")
	}
	u := User{
		Provider: "local",
		Name:     name,
		Email:    email,
		Password: p.hmac.Sum([]byte(password)),
		Token:    email,
		Contact:  Contact{},
	}
	p.db.Create(&u)
	return &u, nil
}

func (p *AuthDao) ResetUid(user uint) {
	p.db.Model(User{}).Where("id = ?", user).Updates(User{Uid: uuid.New()})
}

func (p *AuthDao) Confirm(user uint) {
	now := time.Now()
	p.db.Model(User{}).Where("id = ?", user).Updates(User{Confirmed: &now})
}

func (p *AuthDao) Log(user uint, message string, flag string) {
	p.db.Create(&Log{UserID: user, Message: message, Type: flag})
}

func (p *AuthDao) AddRole(user uint, name string, resource_type string, resource_id uint, startUp, shutDown *time.Time) {
	role := p.getRole(name, resource_type, resource_id)
	for _, pe := range role.Permissions {
		if pe.UserID == user {
			return
		}
	}
	p.db.Create(&Permission{
		UserID: user,
		RoleID: role.ID,
		DateZone: DateZone{
			StartUp:  startUp,
			ShutDown: shutDown,
		},
	})
}

func (p *AuthDao) DelRole(user uint, name string, resource_type string, resource_id uint) {
	role := p.getRole(name, resource_type, resource_id)
	for _, pe := range role.Permissions {
		if pe.UserID == user {
			p.db.Delete(&pe)
			return
		}
	}

}
func (p *AuthDao) Can(user uint, name string, resource_type string, resource_id uint) bool {
	role := p.getRole(name, resource_type, resource_id)
	for _, pe := range role.Permissions {
		if pe.UserID == user {
			return true
		}
	}
	return false
}

func (p *AuthDao) getRole(name string, resource_type string, resource_id uint) *Role {
	role := Role{}
	p.db.FirstOrCreate(&role, Role{Name: name, ResourceType: resource_type, ResourceID: resource_id})

	p.db.Model(&role).Related(&role.Permissions)
	return &role
}
