package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func csrfRequest(method, cookie, header string) bool {
	req := httptest.NewRequest(method, "/x", nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: CSRFCookie, Value: cookie})
	}
	if header != "" {
		req.Header.Set(CSRFHeader, header)
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	return checkCSRF(c)
}

func TestCSRFSafeMethodsNeedNoToken(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		if !csrfRequest(method, "", "") {
			t.Errorf("%s was rejected, but safe methods change nothing and need no token", method)
		}
	}
}

func TestCSRFUnsafeMethodsNeedMatchingPair(t *testing.T) {
	cases := []struct {
		name           string
		cookie, header string
		want           bool
	}{
		{"matching pair", "abc123", "abc123", true},
		{"header missing", "abc123", "", false},
		{"cookie missing", "", "abc123", false},
		{"both missing", "", "", false},
		{"mismatched", "abc123", "def456", false},
		{"header is a prefix", "abc123", "abc", false},
	}

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		for _, tc := range cases {
			if got := csrfRequest(method, tc.cookie, tc.header); got != tc.want {
				t.Errorf("%s with %s: checkCSRF = %v, want %v", method, tc.name, got, tc.want)
			}
		}
	}
}

func TestCookieSecureDefaultsToTrue(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "")
	if !cookieSecure() {
		t.Error("cookieSecure() = false with no setting, want true so production never ships a plain-HTTP cookie")
	}
	t.Setenv("COOKIE_SECURE", "nonsense")
	if !cookieSecure() {
		t.Error("cookieSecure() = false for an unparseable value, want true")
	}
	t.Setenv("COOKIE_SECURE", "false")
	if cookieSecure() {
		t.Error("cookieSecure() = true for an explicit false, want false for local HTTP dev")
	}
}

func TestIssueSessionCookiesSetsHttpOnlySessionAndReadableCSRF(t *testing.T) {
	t.Setenv("COOKIE_DOMAIN", ".cadegray.dev")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)

	if err := IssueSessionCookies(c, "tok-value"); err != nil {
		t.Fatalf("IssueSessionCookies: %v", err)
	}

	got := map[string]*http.Cookie{}
	for _, ck := range rec.Result().Cookies() {
		got[ck.Name] = ck
	}

	session, ok := got[SessionCookie]
	if !ok {
		t.Fatalf("no %s cookie was set", SessionCookie)
	}
	if session.Value != "tok-value" {
		t.Errorf("session cookie value = %q, want the issued token", session.Value)
	}
	if !session.HttpOnly {
		t.Error("session cookie is not HttpOnly, so page scripts could read the token")
	}
	if !session.Secure {
		t.Error("session cookie is not Secure")
	}
	// RFC 6265 says a leading dot is ignored, and Go's parser strips it, so
	// the configured ".cadegray.dev" reads back bare. Either spelling means
	// "cadegray.dev and every subdomain", which is what sharing needs.
	if session.Domain != "cadegray.dev" {
		t.Errorf("session cookie domain = %q, want the shared parent domain", session.Domain)
	}
	if session.SameSite != http.SameSiteLaxMode {
		t.Errorf("session cookie SameSite = %v, want Lax so the post-login redirect carries it", session.SameSite)
	}

	csrf, ok := got[CSRFCookie]
	if !ok {
		t.Fatalf("no %s cookie was set", CSRFCookie)
	}
	if csrf.HttpOnly {
		t.Error("CSRF cookie is HttpOnly, so the page could not echo it back in a header")
	}
	if csrf.Value == "" || csrf.Value == session.Value {
		t.Errorf("CSRF cookie value = %q, want a distinct random value", csrf.Value)
	}
}

func TestClearSessionCookiesExpiresBoth(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/logout", nil)

	ClearSessionCookies(c)

	for _, ck := range rec.Result().Cookies() {
		if ck.MaxAge >= 0 {
			t.Errorf("cookie %s has MaxAge %d, want a negative value to expire it", ck.Name, ck.MaxAge)
		}
	}
}
