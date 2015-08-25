package base

import (
	"github.com/facebookgo/inject"
)

var beans inject.Graph

func LoopEngine(f func(en Engine) error) error {
	for _, obj := range beans.Objects() {
		switch obj.Value.(type) {
		case Engine:
			// en := obj.Value.(Engine)
			// n,v,d := en.Info()
			// log.Printf("%s %s %s", n, v, d)
			if err := f(obj.Value.(Engine)); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

func New(file string) (*Application, error) {
	if cfg, err := Load(file); err == nil {

		return &Application{Cfg: cfg}, nil
	} else {
		return nil, err
	}
}
