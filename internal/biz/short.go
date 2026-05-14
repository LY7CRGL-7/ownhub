package biz

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"ownhub/internal/conf"
)

var ErrShortNotFound = errors.New("short link not found")

type ShortLink struct {
	ID          int64
	Code        string
	URL         string
	Description string
	CreatedAt   time.Time
}

type ShortRepo interface {
	Create(context.Context, *ShortLink) error
	GetByCode(context.Context, string) (*ShortLink, error)
}

type ShortUsecase struct {
	repo          ShortRepo
	publicBaseURL string
}

func NewShortUsecase(repo ShortRepo, app *conf.App) *ShortUsecase {
	baseURL := "http://localhost:8000"
	if app != nil && strings.TrimSpace(app.PublicBaseUrl) != "" {
		baseURL = strings.TrimRight(app.PublicBaseUrl, "/")
	}

	return &ShortUsecase{
		repo:          repo,
		publicBaseURL: baseURL,
	}
}

func (uc *ShortUsecase) Create(ctx context.Context, rawURL, slug, description string) (*ShortLink, string, error) {
	code := strings.TrimSpace(slug)
	if code == "" {
		code = uuid.NewString()[:8]
	}

	entity := &ShortLink{
		Code:        strings.ToLower(code),
		URL:         strings.TrimSpace(rawURL),
		Description: strings.TrimSpace(description),
	}

	if err := uc.repo.Create(ctx, entity); err != nil {
		return nil, "", err
	}

	return entity, uc.publicBaseURL + "/s/" + entity.Code, nil
}

func (uc *ShortUsecase) Get(ctx context.Context, code string) (*ShortLink, string, error) {
	entity, err := uc.repo.GetByCode(ctx, strings.ToLower(strings.TrimSpace(code)))
	if err != nil {
		return nil, "", err
	}
	return entity, uc.publicBaseURL + "/s/" + entity.Code, nil
}
