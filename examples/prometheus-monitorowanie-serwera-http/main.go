package main

import (
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	dec := &decorator{
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "acme",
				Subsystem: "your_app",
				Name:      "http_durations_histogram_seconds",
				Help:      "Request time duration.",
			},
			[]string{"code", "method"},
		),
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "acme",
				Subsystem: "your_app",
				Name:      "http_requests_total",
				Help:      "Total number of requests received.",
			},
			[]string{"code", "method"},
		),
	}

	prometheus.DefaultRegisterer.Register(dec)

	go func() {
		dbg := http.NewServeMux()
		dbg.Handle("/metrics", promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{},
		))
		http.ListenAndServe("0.0.0.0:8081", dbg)
	}()

	app := http.NewServeMux()
	app.Handle("/", dec.instrument(&handler{}))
	http.ListenAndServe("0.0.0.0:8080", app)
}

type handler struct{}

func (h *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	io.WriteString(rw, "It works!")
}

type decorator struct {
	duration *prometheus.HistogramVec
	requests *prometheus.CounterVec
}

func (d *decorator) instrument(handler http.Handler) http.Handler {
	return promhttp.InstrumentHandlerDuration(
		d.duration,
		promhttp.InstrumentHandlerCounter(
			d.requests,
			handler,
		),
	)
}

// Describe implements prometheus.Collector interface.
func (d *decorator) Describe(in chan<- *prometheus.Desc) {
	d.duration.Describe(in)
	d.requests.Describe(in)
}

// Collect implements prometheus.Collector interface.
func (d *decorator) Collect(in chan<- prometheus.Metric) {
	d.duration.Collect(in)
	d.requests.Collect(in)
}
