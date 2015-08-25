package base

import (
	"github.com/op/go-logging"
)

func init() {
	if bkd, err := logging.NewSyslogBackend("itpkg"); err == nil {
		logging.SetBackend(bkd)
	} else {
		log.Error("%v", err)
	}
}
