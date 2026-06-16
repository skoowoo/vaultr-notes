// Package imgfetch implements a server plugin that downloads remote images
// referenced in newly created markdown notes into a configurable vault assets
// directory (default _assets), mirroring the layout used by POST /api/vault/upload-image.
package imgfetch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	urlpkg "net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hardhacker/vaultr/internal/client"
	"github.com/hardhacker/vaultr/internal/config"
	"github.com/hardhacker/vaultr/internal/plugin"
	"github.com/hardhacker/vaultr/internal/storage"
)

const maxImageBytes = 10 << 20

// browserLikeHeaders mimic a desktop Chrome request so CDNs and hotlink
// guards are less likely to reject automated fetches.
const browserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// Plugin implements plugin.Plugin for remote image fetching.
type Plugin struct {
	cfg        config.ImageFetchConfig
	vault      *storage.Vault
	logger     *slog.Logger
	httpClient *http.Client
	eventCh    chan plugin.Event
}

// New creates an image-fetch plugin. The caller must register it only when cfg.Enabled.
func New(cfg config.ImageFetchConfig, vault *storage.Vault, logger *slog.Logger) *Plugin {
	return &Plugin{
		cfg:    cfg,
		vault:  vault,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
		eventCh: make(chan plugin.Event, 128),
	}
}

// Name implements plugin.Plugin.
func (p *Plugin) Name() string { return "image_fetch" }

// Notify implements plugin.Plugin. Non-blocking for create events on .md only.
func (p *Plugin) Notify(e plugin.Event) {
	if e.Type != plugin.EventFSCreate && e.Type != plugin.EventVaultCreate {
		return
	}
	if !strings.HasSuffix(e.Path, ".md") {
		return
	}
	select {
	case p.eventCh <- e:
	default:
		p.logger.Warn("image_fetch: event channel full, dropped", "path", e.Path)
	}
}

// Start implements plugin.Plugin.
func (p *Plugin) Start(ctx context.Context) error {
	p.logger.Info("image_fetch: plugin started", "assets_dir", p.assetsDir())
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-p.eventCh:
			p.handleEvent(ctx, e)
		}
	}
}

// Stop implements plugin.Plugin.
func (p *Plugin) Stop() error { return nil }

func (p *Plugin) assetsDir() string {
	s := strings.Trim(strings.TrimSpace(p.cfg.AssetsDir), "/")
	if s == "" {
		return "_assets"
	}
	return s
}

func (p *Plugin) handleEvent(ctx context.Context, e plugin.Event) {
	notePath, ok := storage.ParsePath(e.Path)
	if !ok {
		p.logger.Info("image_fetch: skip note", "path", e.Path, "reason", "invalid vault path")
		return
	}
	data, err := p.vault.ReadNote(notePath)
	if err != nil {
		p.logger.Info("image_fetch: skip note", "path", e.Path, "reason", "read failed", "err", err)
		return
	}

	urls := client.ParseRemoteImageURLs(data)
	if len(urls) == 0 {
		p.logger.Info("image_fetch: note done", "path", e.Path, "remote_urls", 0)
		return
	}

	p.logger.Info("image_fetch: fetching images for note", "path", e.Path, "remote_urls", len(urls))

	noteStem := strings.TrimSuffix(notePath.Base(), ".md")
	month := time.Now().UTC().Format("200601")
	vaultDir := "/" + path.Join(p.assetsDir(), month)

	var saved, failed int
	for _, u := range urls {
		savedPath, err := p.downloadOne(ctx, u, vaultDir, noteStem)
		if err != nil {
			p.logger.Warn("image_fetch: download failed", "url", u, "note", e.Path, "err", err)
			failed++
			continue
		}
		saved++
		p.logger.Info("image_fetch: saved image", "note", e.Path, "url", u, "image", savedPath)
	}
	p.logger.Info("image_fetch: note done", "path", e.Path, "remote_urls", len(urls), "saved", saved, "failed", failed)
}

func (p *Plugin) downloadOne(ctx context.Context, rawURL, vaultRelDir, noteStem string) (imageVaultPath string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http %d", resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxImageBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", err
	}
	if len(data) > maxImageBytes {
		return "", fmt.Errorf("body exceeds %d bytes", maxImageBytes)
	}

	ext := resolveImageExt(resp.Header.Get("Content-Type"), rawURL)
	if ext == "" {
		return "", fmt.Errorf("could not resolve image extension")
	}
	if !checkImageMagic(data, ext) {
		return "", fmt.Errorf("bytes do not match declared type %s", ext)
	}

	name := imageFilename(noteStem, rawURL, ext)

	rel := strings.TrimPrefix(vaultRelDir, "/")
	absDir := filepath.Join(p.vault.Root(), filepath.FromSlash(rel))
	if err := os.MkdirAll(absDir, 0o750); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(absDir, name), data, 0o644); err != nil {
		return "", err
	}

	img := storage.Image{
		Dir:  vaultRelDir,
		Name: name,
		Ext:  ext,
		Size: int64(len(data)),
	}
	if err := p.vault.RegisterImageWithLinkedNote(img, noteStem); err != nil {
		return "", fmt.Errorf("register metadata: %w", err)
	}
	return storage.JoinPath(vaultRelDir, name), nil
}

// imageFilename builds a deterministic, human-readable filename for a
// downloaded remote image. Format: {noteStem}--{urlSlug}-{urlSig}{ext}
// where urlSig is the first 12 hex chars of SHA-256(rawURL). The sig is
// embedded so BuildImageNoteLinks can reverse-match the file to its source URL.
func imageFilename(noteStem, rawURL, ext string) string {
	noteSlug := sanitizeSlug(noteStem)

	urlSlug := ""
	if pu, err := urlpkg.Parse(rawURL); err == nil {
		base := path.Base(pu.Path)
		if dot := strings.LastIndex(base, "."); dot >= 0 {
			base = base[:dot]
		}
		urlSlug = sanitizeSlug(base)
	}

	h := sha256.Sum256([]byte(rawURL))
	sig := hex.EncodeToString(h[:])[:12]

	switch {
	case noteSlug != "" && urlSlug != "":
		return fmt.Sprintf("%s--%s-%s%s", noteSlug, urlSlug, sig, ext)
	case noteSlug != "":
		return fmt.Sprintf("%s--%s%s", noteSlug, sig, ext)
	case urlSlug != "":
		return fmt.Sprintf("%s-%s%s", urlSlug, sig, ext)
	default:
		return sig + ext
	}
}

// sanitizeSlug converts s to a filesystem-safe slug (ASCII alphanumeric, dash,
// underscore only; spaces become dashes; non-ASCII dropped; max 40 chars).
func sanitizeSlug(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	result := strings.Trim(b.String(), "-")
	if len(result) > 40 {
		result = result[:40]
	}
	return result
}

var imageExtByMIME = map[string]string{
	"image/jpeg":    ".jpg",
	"image/png":     ".png",
	"image/gif":     ".gif",
	"image/webp":    ".webp",
	"image/avif":    ".avif",
	"image/svg+xml": ".svg",
}

var imageMIMEByExt = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".avif": "image/avif",
	".svg":  "image/svg+xml",
}

func resolveImageExt(ct, rawURL string) string {
	ct = strings.ToLower(strings.SplitN(ct, ";", 2)[0])
	if ext, ok := imageExtByMIME[ct]; ok {
		return ext
	}
	pu, err := urlpkg.Parse(rawURL)
	if err != nil {
		return ""
	}
	pathPart := pu.Path
	if pathPart == "" {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(path.Base(pathPart)))
	if _, ok := imageMIMEByExt[ext]; ok {
		return ext
	}
	return ""
}

func checkImageMagic(data []byte, ext string) bool {
	switch ext {
	case ".png":
		return len(data) >= 4 && data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G'
	case ".jpg", ".jpeg":
		return len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8
	case ".gif":
		return len(data) >= 3 && data[0] == 'G' && data[1] == 'I' && data[2] == 'F'
	case ".webp":
		return len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
	default:
		return true
	}
}
