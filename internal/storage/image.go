package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ErrInvalidImageRef indicates dir/name or resolved path is not allowed for gallery delete.
var ErrInvalidImageRef = errors.New("invalid image reference")

// Image is the metadata for a single image file inside a Vault.
type Image struct {
	Dir         string // vault-absolute directory path, e.g. "/_assets/202501" or "/attachments"
	Name        string // filename with extension, e.g. "photo.png"
	Ext         string // lowercase extension including dot, e.g. ".png"
	Size        int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LinkedNotes []string // note basenames (without .md) that embed this image via ![[name]]
}

// wikiImageLinkRe matches Obsidian wiki-style image embeds: ![[filename.ext]] or ![[filename.ext|hint]]
var wikiImageLinkRe = regexp.MustCompile(`!\[\[([^\]\|]+?)(?:\|[^\]]*?)?\]\]`)

// mdImageURLRe captures the URL from standard markdown image syntax: ![alt](url)
var mdImageURLRe = regexp.MustCompile(`!\[[^\]]*\]\((https?://[^)\s]+)\)`)

// imageExtensions is the set of recognised image file extensions (lowercase, with dot).
var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".avif": true,
	".svg":  true,
}

// IsImagePath reports whether a filename has a recognised image extension.
func IsImagePath(name string) bool {
	return imageExtensions[strings.ToLower(filepath.Ext(name))]
}

// GetImagesByName returns all images whose filename equals name.
// Results are ordered by updated_at descending (most recently modified first).
func (g *Vault) GetImagesByName(name string) ([]Image, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbImageGetByName(g.db, name)
}

// RegisterImage inserts or updates the metadata row for an image.
func (g *Vault) RegisterImage(img Image) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return dbImageUpsert(g.db, img)
}

// RegisterImageWithLinkedNote upserts image metadata and sets linked_notes by
// image filename (unique). noteStem is the linking note's stem (without ".md"),
// matching the convention used in BuildImageNoteLinks.
func (g *Vault) RegisterImageWithLinkedNote(img Image, noteStem string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := dbImageUpsert(g.db, img); err != nil {
		return err
	}
	return dbImageSetLinkedNotes(g.db, img.Name, noteStem)
}

// DeleteImageMeta removes the metadata row for the image at (dir, name).
func (g *Vault) DeleteImageMeta(dir, name string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return dbImageDelete(g.db, dir, name)
}

// DeleteImage removes the image file at vault-relative (dir, name) and its
// metadata row. dir is vault-absolute (e.g. "/_assets/202501"); name is the
// basename only. Paths under .vaultr are rejected. Missing files still clear DB.
func (g *Vault) DeleteImage(dir, name string) error {
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return fmt.Errorf("%w: invalid image name %q", ErrInvalidImageRef, name)
	}
	if !IsImagePath(name) {
		return fmt.Errorf("%w: %q is not a recognised image file", ErrInvalidImageRef, name)
	}
	dir = strings.TrimSpace(dir)
	if dir == "" || strings.Contains(dir, "..") {
		return fmt.Errorf("%w: invalid image directory %q", ErrInvalidImageRef, dir)
	}
	if !strings.HasPrefix(dir, "/") {
		dir = "/" + dir
	}
	img := Image{Dir: dir, Name: name}
	abs := g.OsImagePath(img)

	rootWithSep := g.root + string(os.PathSeparator)
	if !strings.HasPrefix(abs, rootWithSep) && abs != g.root {
		return fmt.Errorf("%w: resolved path outside vault", ErrInvalidImageRef)
	}
	internalRoot := filepath.Clean(vaultInternalPath(g.root))
	internalWithSep := internalRoot + string(os.PathSeparator)
	if strings.HasPrefix(abs, internalWithSep) || filepath.Clean(abs) == internalRoot {
		return fmt.Errorf("%w: cannot delete vault-internal path", ErrInvalidImageRef)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if err := os.Remove(abs); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("storage: remove image file: %w", err)
	}
	if err := dbImageDelete(g.db, dir, name); err != nil {
		return fmt.Errorf("storage: delete image metadata: %w", err)
	}
	return nil
}

// ScanAndRegisterImages walks the entire vault root and upserts a metadata row
// for every image file found. Hidden directories (names starting with ".") are
// skipped. Returns the number of images registered.
func (g *Vault) ScanAndRegisterImages() (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if err := dbImageClearAll(g.db); err != nil {
		return 0, err
	}

	count := 0
	err := filepath.WalkDir(g.root, func(absPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !IsImagePath(d.Name()) {
			return nil
		}

		rel, err := filepath.Rel(g.root, absPath)
		if err != nil {
			return nil
		}
		slashRel := "/" + filepath.ToSlash(rel)
		dir, name := PathParts(slashRel)
		ext := strings.ToLower(filepath.Ext(name))

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if dbErr := dbImageUpsert(g.db, Image{
			Dir:       dir,
			Name:      name,
			Ext:       ext,
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
			UpdatedAt: info.ModTime(),
		}); dbErr != nil {
			return nil // best-effort: skip on error
		}
		count++
		return nil
	})
	return count, err
}

// ListImages returns images ordered by updated_at DESC.
// beforeNs is a pagination cursor (Unix nanoseconds); 0 means start from the most recent.
func (g *Vault) ListImages(beforeNs int64, limit int) ([]Image, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbImageListPaged(g.db, beforeNs, limit)
}

// CountImages returns the total number of images in the vault.
func (g *Vault) CountImages() (int, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return dbImageCount(g.db)
}

// BuildImageNoteLinks walks all markdown notes in a single pass, finds every
// ![[image.ext]] wiki embed and every ![alt](https://...) remote image URL,
// and writes the note→image associations back to the DB.
//
// Remote URL associations use the URL fingerprint (first 12 hex chars of
// SHA-256) embedded in filenames by the imgfetch plugin. This lets vaultr init
// reconstruct links even after the DB is wiped, as long as the image files
// retain their fingerprinted names.
//
// The function holds the write lock only during the DB update phase, not during
// the filesystem walk, so it does not block concurrent reads.
func (g *Vault) BuildImageNoteLinks() error {
	// Phase 1: walk all .md files (no lock held — read-only).
	// links: imageName → []noteBasename (from ![[]] embeds)
	// urlSigs: url-sha256-12chars → []noteBasename (from remote image URLs)
	links := make(map[string][]string)
	urlSigs := make(map[string][]string)

	err := filepath.WalkDir(g.root, func(absPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.ToLower(filepath.Ext(d.Name())) != ".md" {
			return nil
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			return nil
		}

		noteName := strings.TrimSuffix(d.Name(), ".md")

		// Match ![[img.ext]] wiki embeds.
		if bytes.Contains(data, []byte("![[")) {
			for _, m := range wikiImageLinkRe.FindAllSubmatch(data, -1) {
				if len(m) < 2 {
					continue
				}
				target := filepath.Base(strings.TrimSpace(string(m[1])))
				if !IsImagePath(target) {
					continue
				}
				links[target] = appendUniqueStr(links[target], noteName)
			}
		}

		// Match ![alt](https://...) remote image URLs; compute URL fingerprint.
		if bytes.Contains(data, []byte("https://")) || bytes.Contains(data, []byte("http://")) {
			for _, m := range mdImageURLRe.FindAllSubmatch(data, -1) {
				if len(m) < 2 {
					continue
				}
				rawURL := strings.TrimSpace(string(m[1]))
				h := sha256.Sum256([]byte(rawURL))
				sig := hex.EncodeToString(h[:])[:12]
				urlSigs[sig] = appendUniqueStr(urlSigs[sig], noteName)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("storage: walk notes for image links: %w", err)
	}

	// Phase 1.5: walk image files; match filenames containing a URL fingerprint
	// to the notes that referenced that URL. This reconstructs imgfetch links.
	if len(urlSigs) > 0 {
		_ = filepath.WalkDir(g.root, func(absPath string, d os.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				if d != nil && d.IsDir() && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if !IsImagePath(d.Name()) {
				return nil
			}
			name := d.Name()
			for sig, noteNames := range urlSigs {
				if strings.Contains(name, sig) {
					for _, n := range noteNames {
						links[name] = appendUniqueStr(links[name], n)
					}
				}
			}
			return nil
		})
	}

	// Phase 2: write results to DB (exclusive lock).
	g.mu.Lock()
	defer g.mu.Unlock()

	if err := dbImageClearAllLinkedNotes(g.db); err != nil {
		return fmt.Errorf("storage: clear image linked notes: %w", err)
	}
	for imgName, noteNames := range links {
		_ = dbImageSetLinkedNotes(g.db, imgName, strings.Join(noteNames, "\n"))
	}
	return nil
}

func appendUniqueStr(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

// OsImagePath resolves the vault-relative Image back to an absolute OS path.
func (g *Vault) OsImagePath(img Image) string {
	rel := img.Dir
	if rel == "/" {
		rel = ""
	}
	return filepath.Join(g.root, filepath.FromSlash(rel), img.Name)
}
