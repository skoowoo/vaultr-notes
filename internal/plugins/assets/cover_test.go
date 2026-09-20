package assets

import (
	"reflect"
	"testing"
)

func TestParseCoverSources(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []coverSource
	}{
		{
			name: "image_url first, then body",
			raw: `---
title: "x"
image_url: "https://cdn.example.com/og.jpg"
---

![body](https://cdn.example.com/body.png)
`,
			want: []coverSource{
				{remote: "https://cdn.example.com/og.jpg"},
				{remote: "https://cdn.example.com/body.png"},
			},
		},
		{
			name: "image_url duplicated in body is listed once",
			raw: `---
image_url: "https://cdn.example.com/og.jpg"
---

![](https://cdn.example.com/og.jpg)
`,
			want: []coverSource{{remote: "https://cdn.example.com/og.jpg"}},
		},
		{
			name: "body images keep document order",
			raw:  "![[a.webp]]\n\n![x](https://cdn.example.com/b.png)\n\n![[c.png]]\n",
			want: []coverSource{
				{local: "a.webp"},
				{remote: "https://cdn.example.com/b.png"},
				{local: "c.png"},
			},
		},
		{
			name: "candidates are capped",
			raw:  "![[a.png]] ![[b.png]] ![[c.png]] ![[d.png]]",
			want: []coverSource{{local: "a.png"}, {local: "b.png"}, {local: "c.png"}},
		},
		{
			name: "serve name query",
			raw:  `![x](/api/images/serve?name=1712-abcd.jpg)`,
			want: []coverSource{{local: "1712-abcd.jpg"}},
		},
		{
			name: "assets path uses basename",
			raw:  `![x](/_assets/202609/1712-abcd.jpg)`,
			want: []coverSource{{local: "1712-abcd.jpg"}},
		},
		{
			name: "invalid first candidate falls through to the next",
			raw: `---
image_url: "data:image/png;base64,aaaa"
---

![](data:image/png;base64,bbbb)

![ok](https://cdn.example.com/ok.png)
`,
			want: []coverSource{{remote: "https://cdn.example.com/ok.png"}},
		},
		{
			name: "images in fenced code are ignored",
			raw:  "```md\n![x](https://cdn.example.com/code.png)\n```\n\n![y](https://cdn.example.com/real.png)\n",
			want: []coverSource{{remote: "https://cdn.example.com/real.png"}},
		},
		{
			name: "url with balanced parens",
			raw:  `![x](https://en.wikipedia.org/wiki/File:A_(b).png "title")`,
			want: []coverSource{{remote: "https://en.wikipedia.org/wiki/File:A_(b).png"}},
		},
		{
			name: "angle-bracket ref with spaces",
			raw:  `![x](<my photo.png>)`,
			want: []coverSource{{local: "my photo.png"}},
		},
		{
			name: "wiki name keeps #",
			raw:  `![[img#1.png|200]]`,
			want: []coverSource{{local: "img#1.png"}},
		},
		{
			name: "protocol-relative url",
			raw:  `![x](//cdn.example.com/a.png)`,
			want: []coverSource{{remote: "https://cdn.example.com/a.png"}},
		},
		{
			name: "non-image and no image",
			raw:  "---\ntitle: \"x\"\n---\n\n![[doc.pdf]]\n\njust text\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCoverSources([]byte(tt.raw))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
