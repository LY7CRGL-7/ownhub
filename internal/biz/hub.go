package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"ownhub/internal/conf"
)

type User struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type NoteItem struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Kind      string    `json:"kind"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	URL       string    `json:"url"`
	Tags      string    `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

type Reminder struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	DueAt     time.Time `json:"due_at"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type UploadAsset struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	FileName    string    `json:"file_name"`
	StoredPath  string    `json:"stored_path"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
}

type WebhookEvent struct {
	ID         int64     `json:"id"`
	Source     string    `json:"source"`
	Event      string    `json:"event"`
	DeliveryID string    `json:"delivery_id"`
	Payload    string    `json:"payload"`
	CreatedAt  time.Time `json:"created_at"`
}

type IPInfo struct {
	Query        string `json:"query"`
	Status       string `json:"status"`
	Country      string `json:"country"`
	RegionName   string `json:"region_name"`
	City         string `json:"city"`
	ISP          string `json:"isp"`
	Organization string `json:"organization"`
	Timezone     string `json:"timezone"`
}

type HubRepo interface {
	CreateUser(context.Context, string, string, string) (*User, error)
	GetUserByUsername(context.Context, string) (*User, string, error)
	GetUserByID(context.Context, int64) (*User, error)
	CreateNote(context.Context, *NoteItem) error
	ListNotes(context.Context, int64, string) ([]*NoteItem, error)
	CreateReminder(context.Context, *Reminder) error
	ListReminders(context.Context, int64, string) ([]*Reminder, error)
	MarkReminderDone(context.Context, int64, int64) error
	CreateUpload(context.Context, *UploadAsset) error
	CreateWebhookEvent(context.Context, *WebhookEvent) error
}

type HubUsecase struct {
	repo          HubRepo
	httpClient    *http.Client
	jwtSecret     []byte
	publicBaseURL string
	uploadDir     string
}

func NewHubUsecase(repo HubRepo, app *conf.App) *HubUsecase {
	baseURL := "http://localhost:8000"
	uploadDir := "./storage/uploads"
	jwtSecret := "change-me-in-production"
	if app != nil {
		if strings.TrimSpace(app.PublicBaseUrl) != "" {
			baseURL = strings.TrimRight(app.PublicBaseUrl, "/")
		}
		if strings.TrimSpace(app.UploadDir) != "" {
			uploadDir = app.UploadDir
		}
		if strings.TrimSpace(app.JwtSecret) != "" {
			jwtSecret = app.JwtSecret
		}
	}

	return &HubUsecase{
		repo:          repo,
		httpClient:    &http.Client{Timeout: 5 * time.Second},
		jwtSecret:     []byte(jwtSecret),
		publicBaseURL: baseURL,
		uploadDir:     uploadDir,
	}
}

func (uc *HubUsecase) Register(ctx context.Context, username, password, displayName string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	return uc.repo.CreateUser(ctx, strings.TrimSpace(username), string(hash), strings.TrimSpace(displayName))
}

func (uc *HubUsecase) Login(ctx context.Context, username, password string) (string, *User, error) {
	user, passwordHash, err := uc.repo.GetUserByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		return "", nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid username or password")
	}

	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(uc.jwtSecret)
	if err != nil {
		return "", nil, fmt.Errorf("sign token: %w", err)
	}
	return signed, user, nil
}

func (uc *HubUsecase) ParseToken(tokenString string) (*User, error) {
	tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer "))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return uc.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	sub, ok := claims["sub"]
	if !ok {
		return nil, fmt.Errorf("missing subject")
	}

	var userID int64
	switch v := sub.(type) {
	case float64:
		userID = int64(v)
	case int64:
		userID = v
	default:
		return nil, fmt.Errorf("invalid subject type")
	}

	return uc.repo.GetUserByID(context.Background(), userID)
}

func (uc *HubUsecase) CreateNote(ctx context.Context, note *NoteItem) error {
	note.Kind = strings.ToLower(strings.TrimSpace(note.Kind))
	if note.Kind == "" {
		note.Kind = "note"
	}
	return uc.repo.CreateNote(ctx, note)
}

func (uc *HubUsecase) ListNotes(ctx context.Context, userID int64, kind string) ([]*NoteItem, error) {
	return uc.repo.ListNotes(ctx, userID, strings.ToLower(strings.TrimSpace(kind)))
}

func (uc *HubUsecase) CreateReminder(ctx context.Context, reminder *Reminder) error {
	if reminder.Status == "" {
		reminder.Status = "pending"
	}
	return uc.repo.CreateReminder(ctx, reminder)
}

func (uc *HubUsecase) ListReminders(ctx context.Context, userID int64, status string) ([]*Reminder, error) {
	return uc.repo.ListReminders(ctx, userID, strings.ToLower(strings.TrimSpace(status)))
}

func (uc *HubUsecase) MarkReminderDone(ctx context.Context, userID, reminderID int64) error {
	return uc.repo.MarkReminderDone(ctx, userID, reminderID)
}

func (uc *HubUsecase) BuildUploadAsset(userID int64, originalName, storedName, contentType string, size int64) *UploadAsset {
	relPath := "/uploads/" + url.PathEscape(storedName)
	return &UploadAsset{
		UserID:      userID,
		FileName:    originalName,
		StoredPath:  filepath.Join(uc.uploadDir, storedName),
		ContentType: contentType,
		Size:        size,
		URL:         uc.publicBaseURL + relPath,
	}
}

func (uc *HubUsecase) SaveUpload(ctx context.Context, asset *UploadAsset) error {
	return uc.repo.CreateUpload(ctx, asset)
}

func (uc *HubUsecase) SaveWebhook(ctx context.Context, source, event, deliveryID string, payload []byte) (*WebhookEvent, error) {
	minified := bytes.TrimSpace(payload)
	var pretty bytes.Buffer
	if json.Valid(minified) {
		if err := json.Indent(&pretty, minified, "", "  "); err == nil {
			minified = pretty.Bytes()
		}
	}

	entity := &WebhookEvent{
		Source:     source,
		Event:      event,
		DeliveryID: deliveryID,
		Payload:    string(minified),
	}
	if err := uc.repo.CreateWebhookEvent(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (uc *HubUsecase) LookupIP(ctx context.Context, ip string) (*IPInfo, error) {
	target := "http://ip-api.com/json/"
	if ip = strings.TrimSpace(ip); ip != "" {
		target += url.PathEscape(ip)
	}
	target += "?fields=status,query,country,regionName,city,timezone,isp,org"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := uc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query ip provider: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ip response: %w", err)
	}
	var raw struct {
		Status     string `json:"status"`
		Query      string `json:"query"`
		Country    string `json:"country"`
		RegionName string `json:"regionName"`
		City       string `json:"city"`
		Timezone   string `json:"timezone"`
		ISP        string `json:"isp"`
		Org        string `json:"org"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode ip response: %w", err)
	}
	return &IPInfo{
		Query:        raw.Query,
		Status:       raw.Status,
		Country:      raw.Country,
		RegionName:   raw.RegionName,
		City:         raw.City,
		ISP:          raw.ISP,
		Organization: raw.Org,
		Timezone:     raw.Timezone,
	}, nil
}
