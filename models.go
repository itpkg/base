package base

import (
	"time"
)

type Model struct {
	ID        uint `gorm:"primary_key"`
	UpdatedAt time.Time
	CreatedAt time.Time `sql:"default:CURRENT_TIMESTAMP;not null"`
}

type VModel struct {
	Uid       string    `sql:"type:UUID;default:;index;not null"`
	Ver       uint      `sql:"default:0;not null"`
	CreatedAt time.Time `sql:"default:CURRENT_TIMESTAMP;not null"`
}

type DateZone struct {
	StartUp  *time.Time `sql:"type:DATE;default:CURRENT_DATE;not null"`
	ShutDown *time.Time `sql:"type:DATE;default:'9999-12-31';not null"`
}

//-----------------------------------------------------------------------------

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

type Setting struct {
	ID  string `gorm:"primary_key"`
	Val []byte `sql:"not null"`
}

type Locale struct {
	ID      uint   `gorm:"primary_key"`
	Lang    string `sql:"index;size:5;not null;default 'en_US'"`
	Code    string `sql:"size:255;index;not null"`
	Message string `sql:"type:TEXT;not null"`
}
