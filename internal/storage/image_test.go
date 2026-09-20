package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

var testPNG = append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 32)...)

func newTestVault(t *testing.T) *Vault {
	t.Helper()
	v, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = v.Close() })
	return v
}

func TestSaveImageBytes(t *testing.T) {
	v := newTestVault(t)

	img, err := v.SaveImageBytes(testPNG, ".PNG")
	if err != nil {
		t.Fatal(err)
	}
	if img.Ext != ".png" || img.Dir == "" || img.Name == "" {
		t.Fatalf("unexpected image %+v", img)
	}
	if _, err := os.Stat(filepath.Join(v.Root(), filepath.FromSlash(img.Dir), img.Name)); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if imgs, _ := v.GetImagesByName(img.Name); len(imgs) != 1 {
		t.Fatalf("image not registered: %v", imgs)
	}

	for name, tc := range map[string]struct {
		data []byte
		ext  string
	}{
		"empty":         {nil, ".png"},
		"bad ext":       {testPNG, ".exe"},
		"magic":         {[]byte("not a png at all"), ".png"},
		"avif not ftyp": {[]byte(`{"secret":true}`), ".avif"},
	} {
		if _, err := v.SaveImageBytes(tc.data, tc.ext); !errors.Is(err, ErrInvalidImageRef) {
			t.Errorf("%s: err = %v, want ErrInvalidImageRef", name, err)
		}
	}
}

func TestSaveImageBytesNamed(t *testing.T) {
	v := newTestVault(t)

	img, err := v.SaveImageBytesNamed(testPNG, "remote-abc123", ".png")
	if err != nil {
		t.Fatal(err)
	}
	if img.Name != "remote-abc123.png" {
		t.Fatalf("name = %q", img.Name)
	}
	if imgs, _ := v.GetImagesByName(img.Name); len(imgs) != 1 {
		t.Fatalf("image not registered: %v", imgs)
	}

	for _, stem := range []string{"", "a/b", `a\b`, "..", "a.b"} {
		if _, err := v.SaveImageBytesNamed(testPNG, stem, ".png"); !errors.Is(err, ErrInvalidImageRef) {
			t.Errorf("stem %q: err = %v, want ErrInvalidImageRef", stem, err)
		}
	}
}

func TestDetectImageExt(t *testing.T) {
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	tests := []struct {
		name string
		data []byte
		ct   string
		file string
		want string
	}{
		{"magic beats wrong content type", jpeg, "image/png", "x.png", ".jpg"},
		{"content type when no magic", []byte("<svg/>"), "image/svg+xml", "", ".svg"},
		{"filename fallback", []byte("<svg/>"), "application/octet-stream", "a.svg", ".svg"},
		{"unknown", []byte("{}"), "application/json", "a.json", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectImageExt(tt.data, tt.ct, tt.file); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeleteImageClearsCoverReference(t *testing.T) {
	v := newTestVault(t)
	img, err := v.SaveImageBytes(testPNG, ".png")
	if err != nil {
		t.Fatal(err)
	}
	sp := Path("/n.md")
	if err := v.ReplaceNoteAssets(sp, AssetKindCover, []NoteAsset{{Kind: AssetKindCover, Filename: img.Name}}); err != nil {
		t.Fatal(err)
	}
	if err := v.DeleteImage(img.Dir, img.Name); err != nil {
		t.Fatal(err)
	}
	_, err = v.GetNoteAsset(sp, AssetKindCover)
	if !errors.Is(err, ErrMetadataNotFound) {
		t.Fatalf("cover after image delete err = %v, want ErrMetadataNotFound", err)
	}
}

func TestNoteAssetFilenamesBatchesAndFiltersMissingImages(t *testing.T) {
	v := newTestVault(t)
	img, err := v.SaveImageBytes(testPNG, ".png")
	if err != nil {
		t.Fatal(err)
	}

	paths := make([]Path, 0, noteAssetLookupBatchSize+50)
	for i := 0; i < noteAssetLookupBatchSize+50; i++ {
		p := Path(JoinPath("/notes", "n"+strconv.Itoa(i)+".md"))
		paths = append(paths, p)
		if err := v.ReplaceNoteAssets(p, AssetKindCover, []NoteAsset{{Kind: AssetKindCover, Filename: img.Name}}); err != nil {
			t.Fatal(err)
		}
	}
	missing := Path("/notes/missing.md")
	paths = append(paths, missing)
	if err := v.ReplaceNoteAssets(missing, AssetKindCover, []NoteAsset{{Kind: AssetKindCover, Filename: "missing.png"}}); err != nil {
		t.Fatal(err)
	}

	got, err := v.NoteAssetFilenames(paths, AssetKindCover)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != noteAssetLookupBatchSize+50 {
		t.Fatalf("cover map size = %d, want %d", len(got), noteAssetLookupBatchSize+50)
	}
	if got[string(missing)] != "" {
		t.Fatalf("missing image should be filtered, got %q", got[string(missing)])
	}
}
