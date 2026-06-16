package util

import (
	"bytes"
	"testing"
)

func TestParseShortNoteFile(t *testing.T) {
	body := []byte(`###### Short Note: 2026-05-10 10:00:00

First line.

---

###### Short Note: 2026-05-10 11:00:00

Second **note**.
`)
	got := ParseShortNoteFile(body)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Timestamp != "2026-05-10 10:00:00" || string(got[0].BodyMD) != "First line." {
		t.Errorf("entry0 = %#v", got[0])
	}
	if got[1].Timestamp != "2026-05-10 11:00:00" || string(got[1].BodyMD) != "Second **note**." {
		t.Errorf("entry1 = %#v", got[1])
	}

	// CRLF delimiter
	body2 := []byte("###### Short Note: t\n\na\r\n---\r\n\r\n###### Short Note: t2\n\nb")
	got2 := ParseShortNoteFile(body2)
	if len(got2) != 2 || string(got2[1].BodyMD) != "b" {
		t.Errorf("crlf split: %#v", got2)
	}
}

func TestParseShortNoteFile_withFrontmatter(t *testing.T) {
	body := []byte("---\nkind: short\ndate: 2026-05-10\n---\n\n###### Short Note: 2026-05-10 10:00:00\n\nFirst.\n\n---\n\n###### Short Note: 2026-05-10 11:00:00\n\nSecond.\n")
	got := ParseShortNoteFile(body)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (frontmatter must be stripped)", len(got))
	}
	if got[0].Timestamp != "2026-05-10 10:00:00" || string(got[0].BodyMD) != "First." {
		t.Errorf("entry0 = %#v", got[0])
	}
	if got[1].Timestamp != "2026-05-10 11:00:00" || string(got[1].BodyMD) != "Second." {
		t.Errorf("entry1 = %#v", got[1])
	}
}

func TestParseShortNoteFile_noHeading(t *testing.T) {
	got := ParseShortNoteFile([]byte("plain **md**"))
	if len(got) != 1 || got[0].Timestamp != "" {
		t.Fatalf("%#v", got)
	}
	if string(got[0].BodyMD) != "plain **md**" {
		t.Fatalf("body %q", got[0].BodyMD)
	}
}

func TestRenderShortNoteFileToHTML_empty(t *testing.T) {
	b, err := RenderShortNoteFileToHTML([]byte(""), nil)
	if err != nil || b != nil {
		t.Fatalf("got %v %v", b, err)
	}
}

func TestRenderShortNoteFileToHTML_escapesTimestamp(t *testing.T) {
	body := []byte(`###### Short Note: <evil>

hi`)
	b, err := RenderShortNoteFileToHTML(body, nil)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(b, []byte("<evil")) {
		t.Fatalf("timestamp not escaped: %s", b)
	}
	if !bytes.Contains(b, []byte("&lt;evil")) {
		t.Fatalf("expected escaped fragment: %s", b)
	}
}

func TestRenderShortNoteFileToHTML_newestFirst(t *testing.T) {
	body := []byte(`###### Short Note: 2026-05-10 10:00:00

Earlier.

---

###### Short Note: 2026-05-10 11:00:00

Later.`)
	b, err := RenderShortNoteFileToHTML(body, nil)
	if err != nil {
		t.Fatal(err)
	}
	iLate := bytes.Index(b, []byte("2026-05-10 11:00:00"))
	iEarly := bytes.Index(b, []byte("2026-05-10 10:00:00"))
	if iLate < 0 || iEarly < 0 {
		t.Fatalf("missing timestamps: %s", b)
	}
	if iLate >= iEarly {
		t.Fatalf("want later timestamp above earlier in DOM order; got late@%d early@%d", iLate, iEarly)
	}
}

func TestRenderShortNoteFileToHTML_withFrontmatter(t *testing.T) {
	body := []byte("---\nkind: short\ndate: 2026-05-10\n---\n\n###### Short Note: 2026-05-10 10:00:00\n\nEarlier.\n\n---\n\n###### Short Note: 2026-05-10 11:00:00\n\nLater.")
	b, err := RenderShortNoteFileToHTML(body, nil)
	if err != nil {
		t.Fatal(err)
	}
	iLate := bytes.Index(b, []byte("2026-05-10 11:00:00"))
	iEarly := bytes.Index(b, []byte("2026-05-10 10:00:00"))
	if iLate < 0 || iEarly < 0 {
		t.Fatalf("missing timestamps: %s", b)
	}
	if iLate >= iEarly {
		t.Fatalf("want later timestamp above earlier in DOM order; got late@%d early@%d", iLate, iEarly)
	}
}
