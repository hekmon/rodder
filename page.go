package rodder

import (
	"fmt"
	"net/http"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

// Page inherits from *rod.Page and adds some useful methods
type Page struct {
	*rod.Page
}

// NewPage creates a new page and optionally apply stealth mode if enabled in the browser
func (b *Browser) NewPage() (page *Page, err error) {
	page = new(Page)
	if b.stealth {
		if page.Page, err = stealth.Page(b.Browser); err != nil {
			err = fmt.Errorf("failed to create a new stealth page: %w", err)
			return
		}
		if err = page.SetWindow(&proto.BrowserBounds{
			WindowState: proto.BrowserWindowStateMaximized,
		}); err != nil {
			err = fmt.Errorf("failed to maximize window: %w", err)
			return
		}
		return
	}
	if page.Page, err = b.Browser.Page(proto.TargetCreateTarget{}); err != nil {
		err = fmt.Errorf("failed to create a new page: %w", err)
		return
	}
	return
}

// GetCookies returns the cookies for the current page
func (p *Page) GetCookies() (cookies []*http.Cookie, err error) {
	rodCookies, err := p.Cookies(nil)
	if err != nil {
		err = fmt.Errorf("failed to extract page cookies: %w", err)
		return
	}
	cookies = make([]*http.Cookie, 0, len(rodCookies))
	for _, rc := range rodCookies {
		cookies = append(cookies, convertRodCookie(rc))
	}
	return
}

// ExtractCookiesTo extracts the cookies from the current page and adds them to the provided jar
func (p *Page) ExtractCookiesTo(jar http.CookieJar) (err error) {
	rodCookies, err := p.Cookies(nil)
	if err != nil {
		err = fmt.Errorf("failed to extract page cookies: %w", err)
		return
	}
	return addCookiesToJar(jar, rodCookies)
}
