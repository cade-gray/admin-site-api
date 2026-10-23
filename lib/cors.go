package lib

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS answers cross-origin requests from an explicit allowlist, with
// credentials enabled so browsers send and accept the session cookie.
//
// The allowlist comes from CORS_ORIGINS (comma-separated, e.g.
// "https://admin.cadegray.dev,https://login.cadegray.dev"). A wildcard is
// deliberately unsupported: the CORS spec forbids pairing
// Access-Control-Allow-Origin: * with Allow-Credentials, so a wildcard would
// silently break cookie auth. Caddy must not add CORS headers of its own.
func CORS(origins string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, o := range strings.Split(origins, ",") {
		if o = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(o), "/")); o != "" {
			allowed[o] = true
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-User, "+CSRFHeader)
			h.Set("Access-Control-Max-Age", "600")
			// The response body varies by Origin, so caches must key on it.
			h.Add("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			// An allowed origin got its headers above; anything else is refused here.
			if origin != "" && !allowed[origin] {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CORSOrigins is the configured allowlist, used for logging at startup.
func CORSOrigins() string {
	return os.Getenv("CORS_ORIGINS")
}
