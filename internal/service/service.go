package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-next-cms/internal/models"
	"go-next-cms/internal/repo"
	"go-next-cms/internal/util"
)

type Service struct{ Repo *repo.Repository }

func New(r *repo.Repository) *Service { return &Service{Repo: r} }

func (s *Service) UniqueSlug(ctx context.Context, countryID int64, title string) (string, error) {
	base := util.Slugify(title)
	slug := base
	for i := 1; i < 100; i++ {
		exists, err := s.Repo.SlugExists(ctx, countryID, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
	return "", errors.New("slug generation failed")
}

func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(s))
}

func AllowedLang(lang string) string {
	if lang == "si" || lang == "ta" || lang == "en" {
		return lang
	}
	return "en"
}

func DealEditable(d *models.Deal, userID int64, isAdmin bool) bool {
	if isAdmin {
		return true
	}
	if d.CreatedByUserID != userID {
		return false
	}
	return d.Status == models.DealPending || d.Status == models.DealRejected
}
