package data

import (
	"context"
	"errors"
	"fmt"

	"ownhub/internal/biz"
)

type hubRepo struct {
	data *Data
}

func NewHubRepo(data *Data) biz.HubRepo {
	return &hubRepo{data: data}
}

func (r *hubRepo) CreateUser(ctx context.Context, username, passwordHash, displayName string) (*biz.User, error) {
	const query = `
INSERT INTO users (username, password_hash, display_name)
VALUES ($1, $2, $3)
RETURNING id, username, display_name, created_at`

	user := &biz.User{}
	err := r.data.db.QueryRowContext(ctx, query, username, passwordHash, displayName).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *hubRepo) GetUserByUsername(ctx context.Context, username string) (*biz.User, string, error) {
	const query = `SELECT id, username, display_name, password_hash, created_at FROM users WHERE username = $1`

	user := &biz.User{}
	var passwordHash string
	err := r.data.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.DisplayName, &passwordHash, &user.CreatedAt,
	)
	if err != nil {
		return nil, "", fmt.Errorf("get user by username: %w", err)
	}
	return user, passwordHash, nil
}

func (r *hubRepo) GetUserByID(ctx context.Context, id int64) (*biz.User, error) {
	const query = `SELECT id, username, display_name, created_at FROM users WHERE id = $1`
	user := &biz.User{}
	err := r.data.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.DisplayName, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *hubRepo) CreateNote(ctx context.Context, note *biz.NoteItem) error {
	const query = `
INSERT INTO notes (user_id, kind, title, content, url, tags)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at`
	return r.data.db.QueryRowContext(ctx, query, note.UserID, note.Kind, note.Title, note.Content, note.URL, note.Tags).
		Scan(&note.ID, &note.CreatedAt)
}

func (r *hubRepo) ListNotes(ctx context.Context, userID int64, kind string) ([]*biz.NoteItem, error) {
	query := `
SELECT id, user_id, kind, title, content, url, tags, created_at
FROM notes
WHERE user_id = $1`
	args := []interface{}{userID}
	if kind != "" {
		query += " AND kind = $2"
		args = append(args, kind)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.data.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*biz.NoteItem
	for rows.Next() {
		item := &biz.NoteItem{}
		if err := rows.Scan(&item.ID, &item.UserID, &item.Kind, &item.Title, &item.Content, &item.URL, &item.Tags, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *hubRepo) CreateReminder(ctx context.Context, reminder *biz.Reminder) error {
	const query = `
INSERT INTO reminders (user_id, title, content, due_at, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at`
	return r.data.db.QueryRowContext(ctx, query, reminder.UserID, reminder.Title, reminder.Content, reminder.DueAt, reminder.Status).
		Scan(&reminder.ID, &reminder.CreatedAt)
}

func (r *hubRepo) ListReminders(ctx context.Context, userID int64, status string) ([]*biz.Reminder, error) {
	query := `
SELECT id, user_id, title, content, due_at, status, created_at
FROM reminders
WHERE user_id = $1`
	args := []interface{}{userID}
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY due_at ASC"

	rows, err := r.data.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*biz.Reminder
	for rows.Next() {
		item := &biz.Reminder{}
		if err := rows.Scan(&item.ID, &item.UserID, &item.Title, &item.Content, &item.DueAt, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *hubRepo) MarkReminderDone(ctx context.Context, userID, reminderID int64) error {
	const query = `UPDATE reminders SET status = 'done' WHERE id = $1 AND user_id = $2`
	result, err := r.data.db.ExecContext(ctx, query, reminderID, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("reminder not found")
	}
	return nil
}

func (r *hubRepo) CreateUpload(ctx context.Context, asset *biz.UploadAsset) error {
	const query = `
INSERT INTO uploads (user_id, file_name, stored_path, content_type, size, public_url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at`
	return r.data.db.QueryRowContext(ctx, query, asset.UserID, asset.FileName, asset.StoredPath, asset.ContentType, asset.Size, asset.URL).
		Scan(&asset.ID, &asset.CreatedAt)
}

func (r *hubRepo) CreateWebhookEvent(ctx context.Context, event *biz.WebhookEvent) error {
	const query = `
INSERT INTO webhook_events (source, event, delivery_id, payload)
VALUES ($1, $2, $3, $4::jsonb)
RETURNING id, created_at`
	return r.data.db.QueryRowContext(ctx, query, event.Source, event.Event, event.DeliveryID, event.Payload).
		Scan(&event.ID, &event.CreatedAt)
}
