package rodder

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

func addCookiesToJar(jar http.CookieJar, rodCookies []*proto.NetworkCookie) (err error) {
	type domainKey struct {
		scheme string
		domain string
	}
	var (
		scheme string
		key    domainKey
		list   []*http.Cookie
		found  bool
	)
	cookiesByDomain := make(map[domainKey][]*http.Cookie, len(rodCookies))
	for _, rodCookie := range rodCookies {
		httpCookie := convertRodCookie(rodCookie)
		if rodCookie.Secure {
			scheme = "https"
		} else {
			scheme = "http"
		}
		key = domainKey{
			scheme: scheme,
			domain: strings.TrimPrefix(rodCookie.Domain, "."), // normalize domain
		}
		if list, found = cookiesByDomain[key]; !found {
			list = make([]*http.Cookie, 0, len(rodCookies))
		}
		cookiesByDomain[key] = append(list, httpCookie)
	}
	var u *url.URL
	for key, cookies := range cookiesByDomain {
		if u, err = url.Parse(fmt.Sprintf("%s://%s", key.scheme, key.domain)); err != nil {
			return fmt.Errorf("failed to parse cookie domain %q: %w", key.domain, err)
		}
		jar.SetCookies(u, cookies)
	}
	return
}

func convertRodCookie(rodCookie *proto.NetworkCookie) (cookie *http.Cookie) {
	cookie = &http.Cookie{
		Name:     rodCookie.Name,
		Value:    rodCookie.Value,
		Path:     rodCookie.Path,
		Domain:   rodCookie.Domain,
		Secure:   rodCookie.Secure,
		HttpOnly: rodCookie.HTTPOnly,
		SameSite: convertSameSite(rodCookie.SameSite),
	}
	if rodCookie.Session {
		cookie.Expires = time.Time{}
		cookie.MaxAge = -1
	} else if rodCookie.Expires > 0 {
		cookie.Expires = time.Unix(int64(rodCookie.Expires), 0)
		cookie.MaxAge = 0 // Rod doesn't provide MaxAge directly
	}
	return
}

func convertSameSite(sameSite proto.NetworkCookieSameSite) http.SameSite {
	switch sameSite {
	case proto.NetworkCookieSameSiteStrict:
		return http.SameSiteStrictMode
	case proto.NetworkCookieSameSiteLax:
		return http.SameSiteLaxMode
	case proto.NetworkCookieSameSiteNone:
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
