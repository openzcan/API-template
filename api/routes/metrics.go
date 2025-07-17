package routes

/*

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myproject_http_requests_total",
			Help: "Total number of HTTP requests by status code and path",
		},
		[]string{"status", "path", "method"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "myproject_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal, requestDuration)
}

// SetupMetricsRoutes configures metrics endpoints
func SetupMetricsRoutes(app *fiber.App) {
	// Dashboard for basic metrics
	app.Get("/metrics/dashboard", monitor.New())

	// Prometheus metrics endpoint
	promHandler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
	app.Get("/metrics", func(c *fiber.Ctx) error {
		promHandler(c.Context())
		return nil
	})

	// Middleware to record metrics
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()

		status := c.Response().StatusCode()
		path := c.Route().Path
		method := c.Method()

		requestsTotal.WithLabelValues(fmt.Sprintf("%d", status), path, method).Inc()
		requestDuration.WithLabelValues(path, method).Observe(duration)

		return err
	})
}
*/
