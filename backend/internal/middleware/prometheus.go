package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/metrics"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//timestamp
		start := time.Now()

		//process
		c.Next()

		duration := time.Since(start).Seconds()
		method := c.Request.Method
		path := c.FullPath()

		if path == "" {
			path = "unknown" //fallback for 404 routes
		}
		status := strconv.Itoa(c.Writer.Status())

		//record telemetry into prometheus
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPReqDuration.WithLabelValues(method, path).Observe(duration)

	}
}
