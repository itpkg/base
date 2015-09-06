package base

import (
	"fmt"

	"github.com/jrallison/go-workers"
	"github.com/op/go-logging"
)

type JobMiddleware struct {
	logger *logging.Logger
}

func (p *JobMiddleware) Call(queue string, message *workers.Msg, next func() bool) (acknowledge bool) {
	p.logger.Info(fmt.Sprintf("BEGIN JOB %s@%s", message.Jid(), queue))
	acknowledge = next()
	p.logger.Info(fmt.Sprintf("END JOB %s@%s", message.Jid(), queue))
	return
}
