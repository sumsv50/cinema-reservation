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

// RequestLogger provides more detailed logging including request/response bodies
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Capture response body
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build full path
		if raw != "" {
			path = path + "?" + raw
		}

		// Log the request
		entry := logrus.WithFields(logrus.Fields{
			"timestamp":     start.Format(time.RFC3339),
			"status_code":   c.Writer.Status(),
			"latency":       latency,
			"client_ip":     c.ClientIP(),
			"method":        c.Request.Method,
			"path":          path,
			"user_agent":    c.Request.UserAgent(),
			"request_size":  c.Request.ContentLength,
			"response_size": c.Writer.Size(),
		})

		// Add request body for POST/PUT requests (be careful with sensitive data)
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if len(requestBody) > 0 && len(requestBody) < 1024 { // Only log small bodies
				entry = entry.WithField("request_body", string(requestBody))
			}
		}

		// Add response body for errors (be careful with sensitive data)
		if c.Writer.Status() >= 400 && w.body.Len() < 1024 {
			entry = entry.WithField("response_body", w.body.String())
		}

		// Log based on status code
		if c.Writer.Status() >= 500 {
			entry.Error("Server Error")
		} else if c.Writer.Status() >= 400 {
			entry.Warn("Client Error")
		} else {
			entry.Info("Request Completed")
		}
	}
}
