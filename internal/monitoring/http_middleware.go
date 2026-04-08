package monitoring

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

func HttpMonitoringMiddleware(ms *PrometheusMetricsService) fiber.Handler {
	return func(c fiber.Ctx) error {

		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()

		ms.ObserveRequestDuration(c.Method(), c.Route().Path, fmt.Sprintf("%d", c.Response().StatusCode()), duration)
		ms.IncrementRequestsTotal(c.Method(), c.Route().Path, fmt.Sprintf("%d", c.Response().StatusCode()))

		return err
	}
}
