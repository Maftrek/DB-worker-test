package prometheus

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
)

type prLogger struct {
	pr *Prometheus
	log.Logger
}

// NewLogger returns new prometheus logger
func NewLogger(l log.Logger, pr *Prometheus) log.Logger {
	return &prLogger{
		Logger: l,
		pr:     pr,
	}
}

func (l *prLogger) Log(keyvals ...interface{}) error {
	for i := 0; i < len(keyvals); i += 2 {
		if val, ok := keyvals[i].(string); ok && val == "err" {
			l.pr.Errs.WithLabelValues().Add(1)
			break
		}
	}

	keyvals = append(keyvals, "caller", l.CallerPrLog(2), "time", time.Now())
	return l.Logger.Log(keyvals...)
}

func (l *prLogger) CallerPrLog(depth int) string {
	_, file, line, _ := runtime.Caller(depth)
	index := strings.LastIndexByte(file, '/') + 1
	return fmt.Sprintf("%s:%d", file[index:], line)
}
