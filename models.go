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
