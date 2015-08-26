package base

import (
	"github.com/jinzhu/gorm"
)

type Engine interface {
	Job()
	Mount()
	Migrate(*gorm.DB)
	Seed(*gorm.DB) error
	Info() (name string, version string, desc string)
}
