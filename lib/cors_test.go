package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

const allowlist = "https://admin.cadegray.dev, https://login.cadegray.dev/"

func corsResponse(t *testing.T, method, origin string) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	router.Use(CORS(allowlist))
	router.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.POST("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(method, "/x", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestCORSAllowsListedOriginWithCredentials(t *testing.T) {
	rec := corsResponse(t, http.MethodGet, "https://admin.cadegray.dev")

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://admin.cadegray.dev" {
		t.Errorf("Allow-Origin = %q, want the request origin echoed back", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Allow-Credentials = %q, want true so the session cookie is sent", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin so caches do not serve one origin's headers to another", got)
	}
}

// A trailing slash in CORS_ORIGINS is a config typo, not a different origin.
func TestCORSNormalizesConfiguredOrigins(t *testing.T) {
	rec := corsResponse(t, http.MethodGet, "https://login.cadegray.dev")

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://login.cadegray.dev" {
		t.Errorf("Allow-Origin = %q, want the origin allowed despite the configured trailing slash", got)
	}
}

func TestCORSNeverSendsWildcard(t *testing.T) {
	for _, origin := range []string{"https://evil.example", "http://admin.cadegray.dev", "https://admin.cadegray.dev.evil.example"} {
		rec := corsResponse(t, http.MethodGet, origin)
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("origin %q got Allow-Origin %q, want no CORS headers at all", origin, got)
		}
	}
}

func TestCORSPreflightRefusesUnlistedOrigin(t *testing.T) {
	if got := corsResponse(t, http.MethodOptions, "https://evil.example").Code; got != http.StatusForbidden {
		t.Errorf("preflight from an unlisted origin = %d, want %d", got, http.StatusForbidden)
	}
	if got := corsResponse(t, http.MethodOptions, "https://admin.cadegray.dev").Code; got != http.StatusNoContent {
		t.Errorf("preflight from a listed origin = %d, want %d", got, http.StatusNoContent)
	}
}

// Server-to-server callers (curl, clique's backend) send no Origin and must
// not be blocked by the browser-facing allowlist.
func TestCORSIgnoresRequestsWithoutOrigin(t *testing.T) {
	rec := corsResponse(t, http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Errorf("no-Origin request = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("no-Origin request got Allow-Origin %q, want none", got)
	}
}
