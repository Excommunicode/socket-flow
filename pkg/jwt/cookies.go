package jwt

import (
	"net/http"

	"github.com/caitlinelfring/go-env-default"
)

var domain = env.GetDefault("DOMAIN", "localhost")

func SetCookie(w http.ResponseWriter, name, value, domain string, maxAge int, secure, httpOnly bool, sameSite http.SameSite) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: sameSite,
	}
	http.SetCookie(w, cookie)
}

func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) error {

	SetCookie(w, "access_token", accessToken, domain, 900, true, true, http.SameSiteLaxMode)
	SetCookie(w, "refresh_token", refreshToken, domain, 86400, true, true, http.SameSiteLaxMode)

	return nil
}

func DeleteAuthCookies(w http.ResponseWriter) error {
	SetCookie(w, "access_token", "", domain, -1, true, true, http.SameSiteLaxMode)
	SetCookie(w, "refresh_token", "", domain, -1, true, true, http.SameSiteLaxMode)

	return nil
}
