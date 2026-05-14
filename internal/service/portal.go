package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"ownhub/internal/biz"
)

type PortalService struct {
	uc *biz.HubUsecase
}

func NewPortalService(uc *biz.HubUsecase) *PortalService {
	return &PortalService{uc: uc}
}

func (s *PortalService) Register(ctx context.Context, username, password, displayName string) (*biz.User, error) {
	return s.uc.Register(ctx, username, password, displayName)
}

func (s *PortalService) Login(ctx context.Context, username, password string) (string, *biz.User, error) {
	return s.uc.Login(ctx, username, password)
}

func (s *PortalService) CurrentUser(authHeader string) (*biz.User, error) {
	return s.uc.ParseToken(authHeader)
}

func (s *PortalService) LookupIP(ctx context.Context, ip string) (*biz.IPInfo, error) {
	return s.uc.LookupIP(ctx, ip)
}

func (s *PortalService) CreateNote(ctx context.Context, note *biz.NoteItem) error {
	return s.uc.CreateNote(ctx, note)
}

func (s *PortalService) ListNotes(ctx context.Context, userID int64, kind string) ([]*biz.NoteItem, error) {
	return s.uc.ListNotes(ctx, userID, kind)
}

func (s *PortalService) CreateReminder(ctx context.Context, reminder *biz.Reminder) error {
	return s.uc.CreateReminder(ctx, reminder)
}

func (s *PortalService) ListReminders(ctx context.Context, userID int64, status string) ([]*biz.Reminder, error) {
	return s.uc.ListReminders(ctx, userID, status)
}

func (s *PortalService) MarkReminderDone(ctx context.Context, userID, reminderID int64) error {
	return s.uc.MarkReminderDone(ctx, userID, reminderID)
}

func (s *PortalService) SaveUploadedFile(ctx context.Context, userID int64, file multipart.File, header *multipart.FileHeader) (*biz.UploadAsset, error) {
	ext := filepath.Ext(header.Filename)
	storedName := uuid.NewString() + ext
	contentType := header.Header.Get("Content-Type")
	asset := s.uc.BuildUploadAsset(userID, header.Filename, storedName, contentType, 0)

	if err := os.MkdirAll(filepath.Dir(asset.StoredPath), 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	dst, err := os.Create(asset.StoredPath)
	if err != nil {
		return nil, fmt.Errorf("create target file: %w", err)
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("save upload file: %w", err)
	}
	asset.Size = size
	if err := s.uc.SaveUpload(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *PortalService) SaveGitHubWebhook(ctx context.Context, secret string, signature string, event string, deliveryID string, payload []byte) (*biz.WebhookEvent, error) {
	if strings.TrimSpace(secret) != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(expected), []byte(signature)) {
			return nil, fmt.Errorf("invalid GitHub signature")
		}
	}
	return s.uc.SaveWebhook(ctx, "github", event, deliveryID, payload)
}

func DecodeJSON(r *http.Request, dest interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dest)
}

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func ParseIDFromPath(path string) (int64, error) {
	chunks := strings.Split(strings.Trim(path, "/"), "/")
	if len(chunks) == 0 {
		return 0, fmt.Errorf("missing id")
	}
	return strconv.ParseInt(chunks[len(chunks)-1], 10, 64)
}

func ParseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(value))
}
