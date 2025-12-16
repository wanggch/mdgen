package note

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var slugRe = regexp.MustCompile(`[^\p{L}\p{N}]+`)

func Slugify(title string) string {
	cleaned := slugRe.ReplaceAllString(strings.TrimSpace(title), "-")
	cleaned = strings.Trim(cleaned, "-")
	cleaned = strings.ReplaceAll(cleaned, "--", "-")
	if len(cleaned) > 80 {
		cleaned = cleaned[:80]
	}
	return cleaned
}

func BuildFileName(t time.Time, title string) string {
	slug := Slugify(title)
	date := t.Format("20060102")
	if slug == "" {
		return fmt.Sprintf("%s.md", date)
	}
	return fmt.Sprintf("%s_%s.md", date, slug)
}

type NoteRequest struct {
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Folder  *string `json:"folder"`
}

type Rendered struct {
	FileName string
	FullPath string
	Content  []byte
}

var ErrInvalidFolder = errors.New("invalid folder")

func ResolveFolder(base string, folder *string) (string, error) {
	target := base
	if folder != nil && strings.TrimSpace(*folder) != "" {
		cleaned := filepath.Clean(strings.TrimSpace(*folder))
		if cleaned == "." {
			// stay in base
		} else if strings.Contains(cleaned, "..") || strings.HasPrefix(cleaned, string(filepath.Separator)) {
			return "", ErrInvalidFolder
		} else {
			target = filepath.Join(base, cleaned)
		}
	}
	return target, nil
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func RenderMarkdown(req NoteRequest, now time.Time) Rendered {
	fileName := BuildFileName(now, req.Title)
	header := fmt.Sprintf("---\ntitle: %s\ncreated_at: %s\n---\n\n", req.Title, now.Format(time.RFC3339))
	body := fmt.Sprintf("# %s\n\n%s\n", req.Title, req.Content)
	return Rendered{
		FileName: fileName,
		Content:  []byte(header + body),
	}
}

func WriteFile(baseDir string, folder *string, rendered Rendered) (Rendered, error) {
	targetDir, err := ResolveFolder(baseDir, folder)
	if err != nil {
		return Rendered{}, err
	}
	if err := EnsureDir(targetDir); err != nil {
		return Rendered{}, fmt.Errorf("create dir: %w", err)
	}

	filePath := filepath.Join(targetDir, rendered.FileName)
	finalPath, err := ensureUniquePath(filePath)
	if err != nil {
		return Rendered{}, fmt.Errorf("prepare path: %w", err)
	}

	tmpPath := finalPath + ".tmp"
	if err := os.WriteFile(tmpPath, rendered.Content, 0o644); err != nil {
		return Rendered{}, fmt.Errorf("write temp: %w", err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return Rendered{}, fmt.Errorf("rename: %w", err)
	}

	rendered.FullPath = finalPath
	rendered.FileName = filepath.Base(finalPath)
	return rendered, nil
}

func ensureUniquePath(path string) (string, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path, nil
	}
	base := strings.TrimSuffix(path, filepath.Ext(path))
	ext := filepath.Ext(path)
	for i := 1; i < 1000; i++ {
		candidate := fmt.Sprintf("%s-%d%s", base, i, ext)
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not find unique filename for %s", path)
}
