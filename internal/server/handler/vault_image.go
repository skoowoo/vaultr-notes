package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hardhacker/vaultr/internal/storage"
)

// UploadImage handles POST /api/vault/upload-image.
// Accepts multipart/form-data with a "file" field containing an image.
// Saves to {vault_root}/_assets/YYYYMM/{unixms}-{rand8}.{ext}.
// Returns {"src": "/_assets/YYYYMM/filename.ext"}.
func (gh *VaultHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "request parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, hdr, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `missing "file" field: `+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := resolveImageExt(hdr.Header.Get("Content-Type"), hdr.Filename)
	if ext == "" {
		http.Error(w, "unsupported image type", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, 10<<20))
	if err != nil {
		http.Error(w, "read error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !checkImageMagic(data, ext) {
		http.Error(w, "file bytes do not match declared image type", http.StatusBadRequest)
		return
	}

	month := time.Now().UTC().Format("200601")
	var randBuf [4]byte
	_, _ = rand.Read(randBuf[:])
	name := fmt.Sprintf("%d-%s%s", time.Now().UnixMilli(), hex.EncodeToString(randBuf[:]), ext)

	absDir := filepath.Join(gh.vault.Root(), "_assets", month)
	if err := os.MkdirAll(absDir, 0o750); err != nil {
		http.Error(w, "mkdir error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(filepath.Join(absDir, name), data, 0o644); err != nil {
		http.Error(w, "write error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Register in images metadata table (best-effort).
	_ = gh.vault.RegisterImage(storage.Image{
		Dir:  "/_assets/" + month,
		Name: name,
		Ext:  ext,
		Size: int64(len(data)),
	})

	respondJSON(w, http.StatusOK, map[string]string{
		"src": "/_assets/" + month + "/" + name,
	})
}

// ServeAsset handles GET /_assets/{path} — streams image files stored in the
// vault's _assets directory. The .vaultr internal directory is always blocked.
func (gh *VaultHandler) ServeAsset(w http.ResponseWriter, r *http.Request) {
	// r.URL.Path is like "/_assets/202501/xxx.png"
	rel := strings.TrimPrefix(r.URL.Path, "/_assets/")
	if rel == "" || strings.Contains(rel, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	root := gh.vault.Root()
	abs := filepath.Join(root, "_assets", filepath.FromSlash(rel))

	// Verify the resolved path stays inside {root}/_assets/.
	assetsRoot := filepath.Join(root, "_assets") + string(os.PathSeparator)
	if !strings.HasPrefix(abs, assetsRoot) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	f, err := os.Open(abs)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "read error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	ct := imageContentType(filepath.Ext(abs))
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, _ = io.Copy(w, f)
}

// ── helpers ───────────────────────────────────────────────────────────────────

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

// resolveImageExt returns the canonical extension for the upload, preferring
// the declared Content-Type and falling back to the filename extension.
// Returns "" when the type is not supported.
func resolveImageExt(ct, filename string) string {
	ct = strings.ToLower(strings.SplitN(ct, ";", 2)[0])
	if ext, ok := imageExtByMIME[ct]; ok {
		return ext
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if _, ok := imageMIMEByExt[ext]; ok {
		return ext
	}
	return ""
}

func imageContentType(ext string) string {
	if ct, ok := imageMIMEByExt[strings.ToLower(ext)]; ok {
		return ct
	}
	return "application/octet-stream"
}

// ServeImageByName handles GET /api/images/serve?name=<filename>.
// It looks up the image by filename in the metadata DB and streams the file.
func (gh *VaultHandler) ServeImageByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}

	imgs, err := gh.vault.GetImagesByName(name)
	if err != nil || len(imgs) == 0 {
		http.NotFound(w, r)
		return
	}

	abs := gh.vault.OsImagePath(imgs[0])
	// Safety check: ensure path stays inside vault root.
	root := gh.vault.Root()
	if !strings.HasPrefix(abs, root+string(os.PathSeparator)) && abs != root {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	f, err := os.Open(abs)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "read error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	ct := imageContentType(filepath.Ext(abs))
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = io.Copy(w, f)
}

// ServeImageAt handles GET /api/images/at?dir=...&name=...
// Serves a specific image by its vault-relative directory and filename.
func (gh *VaultHandler) ServeImageAt(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("dir")
	name := r.URL.Query().Get("name")
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") ||
		strings.Contains(dir, "..") {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	root := gh.vault.Root()
	relDir := strings.TrimPrefix(dir, "/")
	abs := filepath.Join(root, filepath.FromSlash(relDir), name)

	if !strings.HasPrefix(abs, root+string(os.PathSeparator)) && abs != root {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	f, err := os.Open(abs)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "read error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	ct := imageContentType(filepath.Ext(abs))
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = io.Copy(w, f)
}

type imageDeleteRequest struct {
	Dir  string `json:"dir"`
	Name string `json:"name"`
}

// DeleteGalleryImage handles POST /api/images/delete — removes the image file
// and its images table row (dir is vault-absolute, e.g. "/_assets/202501").
func (gh *VaultHandler) DeleteGalleryImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req imageDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := gh.vault.DeleteImage(req.Dir, req.Name); err != nil {
		if errors.Is(err, storage.ErrInvalidImageRef) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeVaultError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// checkImageMagic performs a minimal magic-bytes check for common formats.
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
		// AVIF, SVG — skip deep check
		return true
	}
}
