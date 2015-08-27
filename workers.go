package base

import (
	"fmt"
	"log/syslog"

	"github.com/jrallison/go-workers"
)

type JobMiddleware struct {
	logger *syslog.Writer
}

func (p *JobMiddleware) Call(queue string, message *workers.Msg, next func() bool) (acknowledge bool) {
	p.logger.Info(fmt.Sprintf("BEGIN JOB %s@%s", message.Jid(), queue))
	acknowledge = next()
	p.logger.Info(fmt.Sprintf("END JOB %s@%s", message.Jid(), queue))
	return
}
