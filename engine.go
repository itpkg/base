package base

import (
	"github.com/jinzhu/gorm"
)

type Engine interface {
	Job()
	Mount()
	Migrate(*gorm.DB)
	Seeds(*gorm.DB)
	Info() (name string, version string, desc string)
}
