package base

import (
	"github.com/op/go-logging"
)

func CreateLogger() *logging.Logger {
	return logging.MustGetLogger("itpkg")
}

var log = CreateLogger()
