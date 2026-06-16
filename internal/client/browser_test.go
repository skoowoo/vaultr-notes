package client

import "testing"

func TestViewPageURL(t *testing.T) {
	tests := []struct {
		name       string
		origin     string
		pathOrName string
		want       string
	}{
		{
			name:       "path",
			origin:     "http://127.0.0.1:54321/",
			pathOrName: "dir/note.md",
			want:       "http://127.0.0.1:54321/notes?path=%2Fdir%2Fnote.md",
		},
		{
			name:       "name",
			origin:     "http://127.0.0.1:54321",
			pathOrName: "note.md",
			want:       "http://127.0.0.1:54321/notes?name=note.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ViewPageURL(tt.origin, tt.pathOrName); got != tt.want {
				t.Fatalf("ViewPageURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
