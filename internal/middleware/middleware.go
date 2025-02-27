package middleware

import (
	"github.com/blog/conversational-agent/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware() echo.MiddlewareFunc {
	log := logger.GetLogger()
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:     true,
		LogMethod:  true,
		LogStatus:  true,
		LogLatency: true,
		LogError:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info().
				Str("method", v.Method).
				Str("uri", v.URI).
				Int("status", v.Status).
				Dur("latency", v.Latency).
				Msg("HTTP request processed")
			return nil
		},
	})
}
