package base

import (
	"github.com/go-martini/martini"
)

type Engine interface {
	Job()
	Mount(mrt *martini.ClassicMartini)
	Migrate()
	Seed() error
	Info() (name string, version string, desc string)
}
