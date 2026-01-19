package rodder

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// Browser inherits from *rod.Browser and adds some useful methods
type Browser struct {
	*rod.Browser
	stealth bool
}

var (
	// EnvironmentFR represents the french environment variables to set in the browser process
	EnvironmentFR = []string{
		"TZ=Europe/Paris",
		"LANGUAGE=fr_FR",
		"LC_ALL=fr_FR.UTF-8",
	}
)

// New returns a launched and ready to use browser. It won't use the Leakless side process as it is detected as malware
// by several AV: this means that without a proper call to Close() the browser won't entirely exit and you will need to
// kill it yourself. Stealth mode will disable the headless mode and will add stealth mode to the browser and pages created
// with NewPage() to avoid anti-bot detection.
func New(userProfilDirectory string, stealth bool, additionalEnv []string) (b *Browser, err error) {
	// Create the data dir if necessary
	if err = createDirIfNotExist(userProfilDirectory); err != nil {
		err = fmt.Errorf("failed to create the browser user data dir: %w", err)
		return
	}
	// Create the remote browser
	remoteBrowser := launcher.New()
	remoteBrowser.Leakless(false) // Do not use the leakless side process that is detected as malware by some AV. Close() must be called !
	if stealth {
		remoteBrowser.Headless(false)
		remoteBrowser.Set("enable-automation", "false")
		remoteBrowser.Set("excludeSwitches", "enable-automation")
		remoteBrowser.Set("disable-blink-features", "AutomationControlled")
	}
	remoteBrowser.Set("no-first-run")
	remoteBrowser.Set("no-default-browser-check")
	remoteBrowser.Set("disable-default-apps")
	remoteBrowser.UserDataDir(userProfilDirectory)
	remoteBrowser.Env(
		append(
			os.Environ(),
			additionalEnv...,
		)...,
	)
	// Launch it
	controlURL, err := remoteBrowser.Launch()
	if err != nil {
		err = fmt.Errorf("failed to launch the remote browser: %w", err)
		return
	}
	// Init ourself
	b = &Browser{
		Browser: rod.New().ControlURL(controlURL),
		stealth: stealth,
	}
	if stealth {
		b.Browser.NoDefaultDevice() // maximize view port
	}
	// Connect to the browser dev tools
	if err = b.Browser.Connect(); err != nil {
		err = fmt.Errorf("failed to connect to the remote browser: %w", err)
		return
	}
	return
}

func createDirIfNotExist(path string) (err error) {
	if _, err = os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err = os.MkdirAll(path, 0755); err != nil {
				err = fmt.Errorf("failed to create dir: %w", err)
			}
		} else {
			err = fmt.Errorf("failed to stat dir: %w", err)
		}
	}
	return
}

// ExtractCookiesTo extracts all the cookies from the browser and adds them to the jar
func (b *Browser) ExtractCookiesTo(jar http.CookieJar) (err error) {
	rodCookies, err := b.Browser.GetCookies()
	if err != nil {
		err = fmt.Errorf("failed to extract page cookies: %w", err)
		return
	}
	return addCookiesToJar(jar, rodCookies)
}
