package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ownhub/internal/biz"
)

type shortRepo struct {
	data *Data
}

func NewShortRepo(data *Data) biz.ShortRepo {
	return &shortRepo{data: data}
}

func (r *shortRepo) Create(ctx context.Context, short *biz.ShortLink) error {
	const query = `
INSERT INTO short_links (code, url, description)
VALUES ($1, $2, $3)
RETURNING id, created_at`

	err := r.data.db.QueryRowContext(ctx, query, short.Code, short.URL, short.Description).Scan(&short.ID, &short.CreatedAt)
	if err != nil {
		return fmt.Errorf("create short link: %w", err)
	}
	return nil
}

func (r *shortRepo) GetByCode(ctx context.Context, code string) (*biz.ShortLink, error) {
	const query = `
SELECT id, code, url, description, created_at
FROM short_links
WHERE code = $1`

	var short biz.ShortLink
	err := r.data.db.QueryRowContext(ctx, query, code).Scan(
		&short.ID,
		&short.Code,
		&short.URL,
		&short.Description,
		&short.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, biz.ErrShortNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get short link: %w", err)
	}
	return &short, nil
}
