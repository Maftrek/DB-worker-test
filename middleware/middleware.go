package middleware

import (
	"net/http"
	"strconv"
	"time"

	"DB-worker-test/prometheus"
)

// statusRecorder recodrs response status
type statusRecorder struct {
	http.ResponseWriter
	Status string
}

// WriteHeader reimplements WriteHeader method for ResponseWriter interface
func (rec *statusRecorder) WriteHeader(code int) {
	rec.Status = strconv.Itoa(code)
	rec.ResponseWriter.WriteHeader(code)
}

// Metrics middlware
func Metrics(p *prometheus.Prometheus) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := statusRecorder{ResponseWriter: w}
			now := time.Now()
			next.ServeHTTP(&recorder, r)
			p.Reqs.WithLabelValues(r.Method, recorder.Status, r.URL.Path).Add(1)
			p.Latency.WithLabelValues(r.Method, recorder.Status, r.URL.Path).
				Observe(time.Since(now).Seconds())
		})
	}
}
