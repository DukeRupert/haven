package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func RequestLogger(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			stop := time.Now()

			logger.Info().
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", res.Status).
				Str("ip", c.RealIP()).
				Dur("latency", stop.Sub(start)).
				Msg("Request handled")

			return err
		}
	}
}

func ErrorLogger(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				logger.Error().
					Err(err).
					Str("path", c.Request().URL.Path).
					Str("method", c.Request().Method).
					Msg("Request error occurred")
			}
			return err
		}
	}
}