package compile

import (
	"regexp"
	"strings"
)

// stripEmbeddedBase64DataURIs removes inline RFC 2397 data: URIs that use Base64 encoding:
// metadata includes the ";base64" parameter immediately before the comma that starts the payload.
// Typical forms: markdown images/links, HTML <img src="...">, and <data:...;base64,...>.
// Non-base64 data URLs and normal https images are left unchanged.
func stripEmbeddedBase64DataURIs(s string) string {
	if !strings.Contains(strings.ToLower(s), "data:") {
		return s
	}

	out := reHTMLImgDataBase641.ReplaceAllLiteralString(s, embeddedBase64ImagePlaceholder)
	out = reHTMLImgDataBase642.ReplaceAllLiteralString(out, embeddedBase64ImagePlaceholder)

	out = reMDImgBase64Paren.ReplaceAllStringFunc(out, func(m string) string {
		parts := reMDImgBase64Paren.FindStringSubmatchIndex(m)
		if len(parts) < 4 {
			return mdImageReplacement("")
		}
		alt := m[parts[2]:parts[3]]
		return mdImageReplacement(alt)
	})
	out = reMDImgBase64Angle.ReplaceAllStringFunc(out, func(m string) string {
		parts := reMDImgBase64Angle.FindStringSubmatchIndex(m)
		if len(parts) < 4 {
			return mdImageReplacement("")
		}
		alt := m[parts[2]:parts[3]]
		return mdImageReplacement(alt)
	})

	out = reMDLinkBase64Paren.ReplaceAllStringFunc(out, func(m string) string {
		parts := reMDLinkBase64Paren.FindStringSubmatchIndex(m)
		if len(parts) < 4 {
			return embeddedMDLinkReplacement("")
		}
		text := m[parts[2]:parts[3]]
		return embeddedMDLinkReplacement(text)
	})
	out = reMDLinkBase64Angle.ReplaceAllStringFunc(out, func(m string) string {
		parts := reMDLinkBase64Angle.FindStringSubmatchIndex(m)
		if len(parts) < 4 {
			return embeddedMDLinkReplacement("")
		}
		text := m[parts[2]:parts[3]]
		return embeddedMDLinkReplacement(text)
	})

	out = reAngleBracketDataBase64.ReplaceAllLiteralString(out, embeddedBase64Placeholder)
	return out
}

const (
	embeddedBase64ImagePlaceholder = "*[embedded base64 image omitted]*"
	embeddedBase64Placeholder      = "*[embedded base64 data omitted]*"
)

func mdImageReplacement(alt string) string {
	alt = strings.TrimSpace(alt)
	if alt == "" {
		return embeddedBase64ImagePlaceholder
	}
	return "*[embedded base64 image omitted: " + alt + "]*"
}

func embeddedMDLinkReplacement(linkText string) string {
	linkText = strings.TrimSpace(linkText)
	if linkText == "" {
		return embeddedBase64Placeholder
	}
	return linkText
}

var (
	reMDImgBase64Paren = regexp.MustCompile(`!\[([^\]]*)\]\(\s*(?i:data:[^)]*;base64,[^)]*)\)`)
	reMDImgBase64Angle = regexp.MustCompile(`!\[([^\]]*)\]\(\s*<(?i:data:[^>]*;base64,[^>]*)>\s*\)`)

	// Markdown link to a base64 data URL — keep visible link text, drop the URI.

	reMDLinkBase64Paren = regexp.MustCompile(`\[([^\]]*)\]\(\s*(?i:data:[^)]*;base64,[^)]*)\)`)
	reMDLinkBase64Angle = regexp.MustCompile(`\[([^\]]*)\]\(\s*<(?i:data:[^>]*;base64,[^>]*)>\s*\)`)

	// HTML img with double- or single-quoted src (single-line tags are the common case).

	reHTMLImgDataBase641 = regexp.MustCompile(`(?i)<img\b[^>]*src\s*=\s*"data:[^"]*;base64,[^"]*"[^>]*>`)
	reHTMLImgDataBase642 = regexp.MustCompile(`(?i)<img\b[^>]*src\s*=\s*'data:[^']*;base64,[^']*'[^>]*>`)

	reAngleBracketDataBase64 = regexp.MustCompile(`(?i)<data:[^<>]*;base64,[^<>]*>`)
)
