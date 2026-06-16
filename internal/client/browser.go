package client

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
)

// ViewPageURL builds the browser URL for the server-side HTML view of a note.
// Arguments containing "/" are treated as a vault-relative path; bare names are
// matched vault-wide. origin is the server base URL (e.g. http://127.0.0.1:54321).
func ViewPageURL(origin, pathOrName string) string {
	origin = strings.TrimSuffix(origin, "/")
	q := url.Values{}
	if strings.Contains(pathOrName, "/") {
		p := pathOrName
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		q.Set("path", p)
	} else {
		q.Set("name", pathOrName)
	}
	return origin + "/notes?" + q.Encode()
}

// OpenBrowser opens url in the system default browser.
// Uses platform-appropriate commands: open (macOS), xdg-open (Linux),
// rundll32 url.dll,FileProtocolHandler (Windows).
func OpenBrowser(rawURL string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{rawURL}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", rawURL}
	default: // linux and others
		cmd = "xdg-open"
		args = []string{rawURL}
	}
	if err := exec.Command(cmd, args...).Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
