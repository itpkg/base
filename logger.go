package base

import (
	"fmt"
	"log/syslog"
	"net/http"
	"time"

	"github.com/go-martini/martini"
)

func Logger() martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, ctx martini.Context, log *syslog.Writer) {
		start := time.Now()

		addr := req.Header.Get("X-Real-IP")
		if addr == "" {
			addr = req.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = req.RemoteAddr
			}
		}

		log.Debug(fmt.Sprintf("Started %s %s for %s", req.Method, req.URL.Path, addr))

		rw := res.(martini.ResponseWriter)
		ctx.Next()

		log.Debug(fmt.Sprintf("Completed %v %s in %v\n", rw.Status(), http.StatusText(rw.Status()), time.Since(start)))
	}
}
