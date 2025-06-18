package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func Logger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			// Custom log format
			logrus.WithFields(logrus.Fields{
				"timestamp":   param.TimeStamp.Format(time.RFC3339),
				"status_code": param.StatusCode,
				"latency":     param.Latency,
				"client_ip":   param.ClientIP,
				"method":      param.Method,
				"path":        param.Path,
				"user_agent":  param.Request.UserAgent(),
				"error":       param.ErrorMessage,
			}).Info("HTTP Request")

			return ""
		},
		Output: io.Discard, // We're using logrus, so discard gin's output
	})
}
