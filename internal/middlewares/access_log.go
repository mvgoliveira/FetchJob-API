package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		if status < 400 {
			return
		}

		event := log.Warn()
		if status >= 500 {
			event = log.Error()
		}

		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		event = event.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Int64("latency_ms", time.Since(start).Milliseconds()).
			Str("client_ip", c.ClientIP())

		if errs := c.Errors.ByType(gin.ErrorTypeAny); len(errs) > 0 {
			event = event.Strs("handler_errors", errorMessages(errs))
		}

		event.Msg("http request failed")
	}
}

func errorMessages(errs []*gin.Error) []string {
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}

	return messages
}
