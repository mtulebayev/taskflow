package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by route and status code",
	}, []string{"route", "status"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{.0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1},
	}, []string{"route"})
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (h *Handler) observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		elapsed := time.Since(start)

		httpRequests.WithLabelValues(route, strconv.Itoa(rec.status)).Inc()
		httpDuration.WithLabelValues(route).Observe(elapsed.Seconds())

		switch route {
		case "GET /healthz", "GET /readyz", "GET /metrics":
		default:
			h.log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"route", route,
				"status", rec.status,
				"duration_ms", float64(time.Since(start).Microseconds())/1000,
			)
		}
	})
}
