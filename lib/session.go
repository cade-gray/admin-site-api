package lib

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	// SessionCookie holds the same token /login returns in its body. It is
	// HttpOnly so page scripts can never read it, and scoped to COOKIE_DOMAIN
	// so every app under that domain is signed in by one login.
	SessionCookie = "cg_session"

	// CSRFCookie is readable by scripts on purpose: it is the browser half of
	// the double-submit pair. Its value is meaningless on its own, since
	// proving you can read it is the whole point.
	CSRFCookie = "cg_csrf"

	// CSRFHeader carries the other half, which a cross-site page cannot set.
	CSRFHeader = "X-CSRF-Token"
)

// cookieDomain is the parent domain the session is shared across, e.g.
// ".cadegray.dev". Left empty the cookie is host-only, which is what local
// development over localhost wants.
func cookieDomain() string {
	return os.Getenv("COOKIE_DOMAIN")
}

// cookieSecure defaults to true. Only a local HTTP dev setup should turn it
// off, since browsers refuse to store a Secure cookie over plain HTTP.
func cookieSecure() bool {
	if v := os.Getenv("COOKIE_SECURE"); v != "" {
		secure, err := strconv.ParseBool(v)
		return err != nil || secure
	}
	return true
}

// IssueSessionCookies sets the session and CSRF cookies for a freshly issued
// token. SameSite=Lax is enough here: every app sharing the cookie lives under
// one registrable domain, so its calls to this API are same-site, and the
// post-login hop back from the login page is a top-level GET navigation.
func IssueSessionCookies(c *gin.Context, token string) error {
	csrf, err := randomHex(20)
	if err != nil {
		return err
	}

	c.SetSameSite(http.SameSiteLaxMode)
	// MaxAge 0 makes these session cookies: they last as long as the browser
	// stays open, which matches the non-expiring tokens in auth_db.
	c.SetCookie(SessionCookie, token, 0, "/", cookieDomain(), cookieSecure(), true)
	c.SetCookie(CSRFCookie, csrf, 0, "/", cookieDomain(), cookieSecure(), false)
	return nil
}

// ClearSessionCookies expires both cookies, so a logout on one app signs the
// browser out of all of them.
func ClearSessionCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookie, "", -1, "/", cookieDomain(), cookieSecure(), true)
	c.SetCookie(CSRFCookie, "", -1, "/", cookieDomain(), cookieSecure(), false)
}

// checkCSRF compares the X-CSRF-Token header against the cg_csrf cookie for
// requests that change state. Only cookie-authenticated requests need this: a
// cross-site page can make the browser send cookies, but it cannot read them,
// and it cannot set an Authorization header without clearing CORS preflight
// against the origin allowlist first.
func checkCSRF(c *gin.Context) bool {
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}

	cookie, err := c.Cookie(CSRFCookie)
	if err != nil || cookie == "" {
		return false
	}
	header := c.GetHeader(CSRFHeader)
	return subtle.ConstantTimeCompare([]byte(cookie), []byte(header)) == 1
}

// userForToken returns the username a token belongs to, for cookie auth where
// the client never tells us who it claims to be.
func userForToken(db *gorm.DB, tokenString string) (string, bool) {
	var record Token
	if err := db.Where("token = ?", tokenString).First(&record).Error; err != nil {
		return "", false
	}
	return record.Username, true
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
