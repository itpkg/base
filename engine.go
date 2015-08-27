package base

import (
	"github.com/go-martini/martini"
	"github.com/jrallison/go-workers"
)

type Engine interface {
	Cron()
	Job() (string, func(message *workers.Msg), float32)
	Mount(mrt *martini.ClassicMartini)
	Migrate()
	Seed() error
	Info() (name string, version string, desc string)
}
