package storage

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LocalUploader struct {
	Dir      string
	BasePath string
	MaxSize  int64
}

func (u LocalUploader) Save(c *fiber.Ctx, field string) (string, error) {
	f, err := c.FormFile(field)
	if err != nil {
		return "", nil
	}
	if f.Size > u.MaxSize {
		return "", fmt.Errorf("file too large")
	}
	if !isAllowed(f.Filename) {
		return "", fmt.Errorf("invalid file type")
	}
	name := fmt.Sprintf("%d-%s", time.Now().UnixNano(), sanitize(f.Filename))
	path := filepath.Join(u.Dir, name)
	if err := os.MkdirAll(u.Dir, 0o755); err != nil {
		return "", err
	}
	if err := c.SaveFile(f, path); err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join(u.BasePath, name)), nil
}

func isAllowed(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}

func sanitize(name string) string { return strings.ReplaceAll(filepath.Base(name), " ", "-") }

func _(_ *multipart.FileHeader) {}
