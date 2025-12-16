package note

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"Hello World!", "Hello-World"},
		{"  中文 标题 ", "中文-标题"},
		{strings.Repeat("a", 100), strings.Repeat("a", 80)},
		{"multi--dash###", "multi-dash"},
	}

	for _, tt := range tests {
		if got := Slugify(tt.title); got != tt.want {
			t.Fatalf("slugify(%q) = %q, want %q", tt.title, got, tt.want)
		}
	}
}

func TestEnsureUniquePath(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "20230101_title.md")

	if _, err := ensureUniquePath(base); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.WriteFile(base, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	path, err := ensureUniquePath(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(path, "-1.md") {
		t.Fatalf("expected suffixed filename, got %s", path)
	}
}

func TestResolveFolder(t *testing.T) {
	base := "/tmp/base"
	tests := []struct {
		folder   *string
		wantPath string
		wantErr  bool
	}{
		{nil, base, false},
		{ptr("a/b"), "/tmp/base/a/b", false},
		{ptr("../etc"), "", true},
		{ptr("/abs"), "", true},
	}

	for _, tt := range tests {
		got, err := ResolveFolder(base, tt.folder)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("expected error for folder %v", tt.folder)
			}
			continue
		}
		if err != nil || got != tt.wantPath {
			t.Fatalf("got %s err %v, want %s", got, err, tt.wantPath)
		}
	}
}

func TestRenderMarkdown(t *testing.T) {
	req := NoteRequest{Title: "Example", Content: "Body"}
	now := time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC)
	rendered := RenderMarkdown(req, now)
	expected := "---\ntitle: Example\ncreated_at: 2023-01-02T03:04:05Z\n---\n\n# Example\n\nBody\n"
	if string(rendered.Content) != expected {
		t.Fatalf("unexpected markdown:\n%s", rendered.Content)
	}
	if rendered.FileName != "20230102_Example.md" {
		t.Fatalf("unexpected filename: %s", rendered.FileName)
	}
}

func ptr(s string) *string { return &s }
