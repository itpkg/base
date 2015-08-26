package base

import (
	"time"
)

type Model struct {
	ID        uint `gorm:"primary_key"`
	UpdatedAt time.Time
	CreatedAt time.Time `sql:"type:DATETIME;default:CURRENT_TIMESTAMP"`
}

type VModel struct {
	Uid       string    `sql:"type:UUID;default:;index;not null"`
	Ver       uint      `sql:"default:0;not null"`
	CreatedAt time.Time `sql:"type:DATETIME;default:CURRENT_TIMESTAMP"`
}

type DateZone struct {
	StartUp  *time.Time `sql:"type:DATE;default:CURRENT_DATE"`
	ShutDown *time.Time `sql:"type:DATE;default:'9999-12-31'"`
}
