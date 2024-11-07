package logger

import (
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Logger struct {
	zlog zerolog.Logger
}

// Initialize creates a new configured Logger
func Initialize(environment string) *Logger {
	// Set up console writer with color and human-friendly format for development
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    environment == "production",
	}

	// Configure global settings
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	
	// Set appropriate log level based on environment
	logLevel := zerolog.InfoLevel
	if environment == "development" {
		logLevel = zerolog.DebugLevel
	}

	zlog := zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{
		zlog: zlog,
	}
}

// Middleware creates a new middleware for Echo
func (l *Logger) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Request completion details
			req := c.Request()
			res := c.Response()

			// Log fields
			event := l.zlog.Info().
				Str("method", req.Method).
				Str("uri", req.RequestURI).
				Str("path", req.URL.Path).
				Int("status", res.Status).
				Str("ip", c.RealIP()).
				Dur("latency", time.Since(start)).
				Str("user_agent", req.UserAgent())

			// Add error if present
			if err != nil {
				event = event.Err(err)
			}

			// Add request ID if present
			if reqID := req.Header.Get("X-Request-ID"); reqID != "" {
				event = event.Str("request_id", reqID)
			}

			// Log the request
			event.Msg("Request processed")

			return err
		}
	}
}

// Logger interface methods
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	event := l.zlog.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *Logger) Info(msg string, fields map[string]interface{}) {
	event := l.zlog.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	event := l.zlog.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	event := l.zlog.Error().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *Logger) Fatal(msg string, err error, fields map[string]interface{}) {
	event := l.zlog.Fatal().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}