package base

type Engine interface {
	Job()
	Mount()
	Migrate()
	Info() (name string, version string, desc string)
}
