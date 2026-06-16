package wechat

import (
	"strings"
	"testing"
)

func TestExtractTextPlain(t *testing.T) {
	msg := Message{ItemList: []MessageItem{{
		Type:     MessageItemTypeText,
		TextItem: &TextItem{Text: "hello"},
	}}}
	if got := ExtractText(msg); got != "hello" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractTextQuote(t *testing.T) {
	msg := Message{ItemList: []MessageItem{{
		Type: MessageItemTypeText,
		TextItem: &TextItem{Text: "reply"},
		RefMsg: &RefMessage{
			Title: "title",
			MessageItem: &MessageItem{
				TextItem: &TextItem{Text: "quoted"},
			},
		},
	}}}
	got := ExtractText(msg)
	if !strings.Contains(got, "[引用:") || !strings.Contains(got, "reply") {
		t.Fatalf("got %q", got)
	}
}

func TestSplitText(t *testing.T) {
	segs := SplitText("aaaa\nbbbb", 5)
	if len(segs) != 2 || segs[0] != "aaaa" || segs[1] != "bbbb" {
		t.Fatalf("%v", segs)
	}
}

func TestFormatForWeChat(t *testing.T) {
	in := "**bold** and [link](https://x.com)"
	got := FormatForWeChat(in)
	if strings.Contains(got, "**") {
		t.Fatalf("got %q", got)
	}
	if !strings.Contains(got, "link (https://x.com)") {
		t.Fatalf("got %q", got)
	}
}
