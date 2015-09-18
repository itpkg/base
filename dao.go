package base

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/itpkg/web"
	"github.com/jinzhu/gorm"
	"github.com/magiconair/properties"
	"github.com/op/go-logging"
)

type AuthDao struct {
	HMac *web.HMac `inject:""`
	Aes  *web.Aes  `inject:""`
}

func (p *AuthDao) Auth(db *gorm.DB, email, password string) *User {
	if user := p.GetByEmail(db, email); user != nil && p.HMac.Equal([]byte(password), user.Password) {
		return user
	}
	return nil
}

func (p *AuthDao) GetByEmail(db *gorm.DB, email string) *User {
	var user User
	if db.Where("email = ? AND provider = ?", email, "email").First(&user).RecordNotFound() {
		return nil
	}
	return &user
}

func (p *AuthDao) CreateByEmail(db *gorm.DB, email, name, password string) *User {
	u := User{
		Provider: "email",
		Name:     name,
		Email:    email,
		Password: p.HMac.Sum([]byte(password)),
		Token:    email,
		Contact:  Contact{},
	}
	db.Create(&u)
	return &u
}

func (p *AuthDao) ResetUid(db *gorm.DB, user uint) {
	db.Model(User{}).Where("id = ?", user).Updates(User{Uid: web.Uuid()})
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

//-----------------------------------------------------------------------------
type LocaleDao struct {
	Logger *logging.Logger `inject:""`
}

func (p *LocaleDao) T(db *gorm.DB, lang, code string, args ...interface{}) string {
	val := p.Get(db, lang, code)
	if val == "" {
		return fmt.Sprintf("Translation [%s] not found", code)
	} else {
		return fmt.Sprintf(val, args...)
	}
}

func (p *LocaleDao) Load(db *gorm.DB, path string) error {
	if files, err := ioutil.ReadDir(path); err == nil {
		for _, f := range files {
			fn := f.Name()
			p.Logger.Info("Find %s/%s", path, fn)
			prop := properties.MustLoadFile(path+"/"+fn, properties.UTF8)
			for _, k := range prop.Keys() {
				p.Set(db, k[0:5], k[6:len(k)], prop.MustGetString(k))
			}
		}
		return nil
	} else {
		return err
	}
}

func (p *LocaleDao) Set(db *gorm.DB, lang, code, message string) {
	var l Locale
	if db.Select("id").Where("lang = ? AND code = ?", lang, code).First(&l).RecordNotFound() {
		db.Create(&Locale{Lang: lang, Code: code, Message: message})
	} else {
		db.Model(Locale{}).Where("id = ?", l.ID).Updates(Locale{Message: message})
	}
}

func (p *LocaleDao) Get(db *gorm.DB, lang, code string) string {
	l := Locale{}
	db.Select("message").Where("lang = ? AND code = ?", lang, code).First(&l)
	return l.Message
}

//-----------------------------------------------------------------------------
type SettingDao struct {
	Aes *web.Aes `inject:""`
}

func (p *SettingDao) Set(db *gorm.DB, key string, val interface{}, enc bool) error {
	dt, err := web.ToBits(val)
	if err != nil {
		db.Rollback()
		return err
	}

	if enc {
		dt, err = p.Aes.Encode(dt)
		if err != nil {
			db.Rollback()
			return err
		}
	}

	st := Setting{ID: key}
	var cn int
	db.Model(st).Count(&cn)
	if cn == 0 {
		st.Val = dt
		db.Create(&st)
	} else {
		db.Model(&st).Updates(Setting{Val: dt})
	}
	return nil
}

func (p *SettingDao) Get(db *gorm.DB, key string, val interface{}, enc bool) error {
	st := Setting{}
	db.Where("id = ?", key).First(&st)
	if st.Val != nil {
		var dt []byte

		if enc {
			dt = p.Aes.Decode(st.Val)
		} else {
			dt = st.Val
		}
		return web.FromBits(dt, val)
	}
	return nil
}
