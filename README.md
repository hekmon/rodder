# rodder

[![Go Reference](https://pkg.go.dev/badge/github.com/hekmon/rodder.svg)](https://pkg.go.dev/github.com/hekmon/rodder)

Wrapper around [go-rod](https://github.com/go-rod/rod). It simplifies the boilerplating for creating a remote controlled browser with or without stealth mode. In addition, it provides helper methods to retreive the headers sent by the browser and current cookies from the browser or the current page. This way you can easily setup a http client that mimics a real browser with a initialized cookie jar (for example passing a login page with cloudfare protection before scripting crawling or scraping).

Leakless mod is disable because a lot of AVs detect it as malware. This only means you have to ensure to call `Close()` on the instanciated browser to properly close it.

Every objects inherits from rod's original objects:

* Every original methods are available
* Special (new) helpers methods are added on top of it (check doc)

## Download

```bash
go get -u github.com/hekmon/rodder
```

## Example

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "net/http/cookiejar"
    "os"
    "path/filepath"
    "time"

    "github.com/hashicorp/go-cleanhttp"
    "github.com/hekmon/rodder"
    "golang.org/x/net/publicsuffix"
)

func main() {
    // Create a user profil dir to keep sessions between launchs
    userCacheDir, err := os.UserCacheDir()
    if err != nil {
        panic(err)
    }
    profilDir := filepath.Join(userCacheDir, "project-roduserdata")

    // Create and launch the browser (additional env can be nil)
    stealth := true // Activate all the necessary flags and JS injection on page to have anti-bot bypass
    browser, err := rodder.New(profilDir, stealth, rodder.EnvironmentFR)
    if err != nil {
        panic(err)
    }
    defer browser.Close() // IMPORTANT: otherwise some background chromium process will remain even with all pages are closed !

    // Create a stealth page (high level wrapper that inject rod JS stealth script at startup)
    page, err := browser.NewPage()
    if err != nil {
        panic(err)
    }
    defer page.Close() // not necessary but good habit (think closing the browser tab)

    // Use rod as usual (perform login, pass cloudflare challenge, etc... the idea is the set all the needed cookies here)
    if err = page.Navigate("https://bot.sannysoft.com/"); err != nil {
        panic(err)
    }
    if err = page.WaitStable(time.Second); err != nil {
        panic(err)
    }
    time.Sleep(5 * time.Second) // give you the time to check the page within the open browser

    // Let's create a http client that mimics the browser (except the TLS fingerprinting!)
    //// Let's retreive the current browser headers fingerprint (it will open and close a new tab on the remote browser)
    headers, err := browser.GetHeaders()
    if err != nil {
        panic(err)
    }
    //// Create the custom client
    cookiesClient := cleanhttp.DefaultPooledClient()
    //// Associate the client a cookie jar and inject the current page cookies to it
    jar, err := cookiejar.New(&cookiejar.Options{
        PublicSuffixList: publicsuffix.List,
    })
    if err != nil {
        panic(err)
    }
    if err = page.ExtractCookiesTo(jar); err != nil {
        panic(err)
    }
    cookiesClient.Jar = jar
    //// Create a mimic request with the same browser headers
    req, err := http.NewRequest("GET", "https://bot.sannysoft.com/", nil)
    if err != nil {
        panic(err)
    }
    for header, values := range headers {
        for _, value := range values {
            req.Header.Set(header, value)
        }
    }
    //// Execute it and take advantage of our initialized cookie jar
    resp, err := cookiesClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    //// Enjoy
    fmt.Println(resp.Status)
    fmt.Println()
    data, err := io.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(data))
}
```
