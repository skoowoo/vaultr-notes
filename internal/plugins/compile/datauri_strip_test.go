package compile

import (
	"strings"
	"testing"
)

func TestStripEmbeddedBase64DataURIs_markdownImages(t *testing.T) {
	b64Payload := strings.Repeat("A", 200)
	tests := []struct {
		name     string
		input    string
		contains string
		omit     string
	}{
		{
			name:     "svg with alt",
			input:    "hi\n\n![Agent loop](DATA:image/svg+xml;base64," + b64Payload + ")\nbye",
			contains: "*[embedded base64 image omitted: Agent loop]*",
			omit:     b64Payload,
		},
		{
			name:     "no alt",
			input:    "![](data:image/png;BASE64," + b64Payload + ")",
			contains: embeddedBase64ImagePlaceholder,
			omit:     b64Payload,
		},
		{
			name:     "angle markdown url",
			input:    "![](" + "<DATA:application/octet-stream;" + strings.ToUpper("base64") + "," + b64Payload + ">)",
			contains: embeddedBase64ImagePlaceholder,
			omit:     b64Payload,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := stripEmbeddedBase64DataURIs(tc.input)
			if !strings.Contains(out, tc.contains) {
				t.Fatalf("want substring %q\n got %q", tc.contains, out)
			}
			if strings.Contains(out, tc.omit) {
				t.Fatalf("payload should be stripped, got snippet of payload in:\n%s", out)
			}
		})
	}
}

func TestStripEmbeddedBase64DataURIs_preservesNonBase64DataURI(t *testing.T) {
	// Percent-encoded plaintext data URI (no ";base64" before comma)
	in := `[x](DATA:,hello%29world)`
	got := stripEmbeddedBase64DataURIs(in)
	if got != in {
		t.Fatalf("want unchanged, got %q", got)
	}
}

func TestStripEmbeddedBase64DataURIs_externalImage(t *testing.T) {
	in := `![x](https://example.com/z.png)`
	got := stripEmbeddedBase64DataURIs(in)
	if got != in {
		t.Fatalf("want unchanged, got %q", got)
	}
}

func TestStripEmbeddedBase64DataURIs_mdLinkKeepsLabel(t *testing.T) {
	b64Payload := strings.Repeat("x", 100)
	in := `[click here](DATA:image/png;bAsE64,` + b64Payload + `)`
	want := `click here`
	got := stripEmbeddedBase64DataURIs(in)
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestStripEmbeddedBase64DataURIs_HTMLImg(t *testing.T) {
	b64Payload := strings.Repeat("y", 100)
	in := `<p><IMG class="z" Src="DaTa:image/jpeg;BaSe64,` + b64Payload + `"></p>`
	got := stripEmbeddedBase64DataURIs(in)
	if strings.Contains(got, b64Payload) {
		t.Fatalf("stripped payload should be gone:\n%s", got)
	}
	if !strings.Contains(got, embeddedBase64ImagePlaceholder) {
		t.Fatalf("want placeholder in %q", got)
	}
}

func TestStripEmbeddedBase64DataURIs_angleAutolink(t *testing.T) {
	b64Payload := strings.Repeat("z", 50)
	in := `see <DaTa:image/gif;bAsE64,` + b64Payload + "> ok"
	got := stripEmbeddedBase64DataURIs(in)
	if strings.Contains(got, b64Payload) {
		t.Fatalf("stripped payload should be gone:\n%s", got)
	}
	if !strings.Contains(got, embeddedBase64Placeholder) {
		t.Fatalf("want placeholder in %q", got)
	}
}
