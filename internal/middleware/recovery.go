package middleware

import (
	"net"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"cinema-reservation/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Check for a broken connection, as it is not really a
		// condition that warrants a panic stack trace.
		var brokenPipe bool
		if ne, ok := recovered.(*net.OpError); ok {
			if se, ok := ne.Err.(*os.SyscallError); ok {
				if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
					strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
					brokenPipe = true
				}
			}
		}

		httpRequest, _ := httputil.DumpRequest(c.Request, false)

		if brokenPipe {
			logrus.WithFields(logrus.Fields{
				"error":   recovered,
				"request": string(httpRequest),
			}).Error("Broken pipe detected")

			// If the connection is dead, we can't write a status to it.
			c.Error(recovered.(error))
			c.Abort()
			return
		}

		// Log the panic with stack trace
		logrus.WithFields(logrus.Fields{
			"error":      recovered,
			"request":    string(httpRequest),
			"stack":      string(debug.Stack()),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}).Error("Panic recovered")

		// Return a generic error response
		utils.ErrorResponse(c, nil)
	})
}
