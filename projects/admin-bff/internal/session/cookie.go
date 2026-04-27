package session

import (
	"net/http"
	"time"
)

const (
	CookieName     = "admin_session"
	CSRFCookieName = "admin_csrf"
)

// SetCookies — устанавливает session + csrf cookies на response.
//
// admin_session: HttpOnly Secure SameSite=Strict — недоступен JS.
// admin_csrf:    Secure SameSite=Strict, читается JS — для double-submit.
func SetCookies(w http.ResponseWriter, sessionID, csrfToken, domain string, secure bool, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    sessionID,
		Path:     "/",
		Domain:   domain,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(maxAge.Seconds()),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    csrfToken,
		Path:     "/",
		Domain:   domain,
		HttpOnly: false, // JS читает для X-CSRF-Token header.
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(maxAge.Seconds()),
	})
}

// ClearCookies — удаляет session + csrf cookies (logout).
func ClearCookies(w http.ResponseWriter, domain string, secure bool) {
	for _, name := range []string{CookieName, CSRFCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			Domain:   domain,
			HttpOnly: name == CookieName,
			Secure:   secure,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
		})
	}
}

// SessionIDFromRequest — читает session ID из cookie.
func SessionIDFromRequest(r *http.Request) string {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
