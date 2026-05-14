package biz

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"ownhub/internal/conf"
)

const tokenAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type ToolUsecase struct {
	appName string
}

func NewToolUsecase(app *conf.App) *ToolUsecase {
	name := "personal-backend-hub"
	if app != nil && strings.TrimSpace(app.Name) != "" {
		name = app.Name
	}
	return &ToolUsecase{appName: name}
}

func (uc *ToolUsecase) Ping(context.Context) (string, string) {
	return uc.appName, time.Now().Format(time.RFC3339)
}

func (uc *ToolUsecase) Timestamp(_ context.Context, timezone string) (time.Time, *time.Location, error) {
	loc := time.Local
	if tz := strings.TrimSpace(timezone); tz != "" {
		loaded, err := time.LoadLocation(tz)
		if err != nil {
			return time.Time{}, nil, fmt.Errorf("load timezone: %w", err)
		}
		loc = loaded
	}

	now := time.Now().In(loc)
	return now, loc, nil
}

func (uc *ToolUsecase) GenerateToken(_ context.Context, length int32) (string, int32, error) {
	if length <= 0 {
		length = 16
	}
	if length > 128 {
		length = 128
	}

	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", 0, fmt.Errorf("generate random bytes: %w", err)
	}

	alphabetLen := byte(len(tokenAlphabet))
	for i := range buf {
		buf[i] = tokenAlphabet[int(buf[i]%alphabetLen)]
	}

	return string(buf), length, nil
}

func (uc *ToolUsecase) DigestText(_ context.Context, text, algorithm string) (string, string, error) {
	algo := strings.ToLower(strings.TrimSpace(algorithm))
	if algo == "" {
		algo = "sha256"
	}

	switch algo {
	case "md5":
		sum := md5.Sum([]byte(text))
		return algo, hex.EncodeToString(sum[:]), nil
	case "sha1":
		sum := sha1.Sum([]byte(text))
		return algo, hex.EncodeToString(sum[:]), nil
	case "sha256":
		sum := sha256.Sum256([]byte(text))
		return algo, hex.EncodeToString(sum[:]), nil
	case "sha512":
		sum := sha512.Sum512([]byte(text))
		return algo, hex.EncodeToString(sum[:]), nil
	default:
		return "", "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// ========== AI 方法 ==========

// ChatMessage 对话消息
type ChatMessage struct {
	Role    string
	Content string
}

// AIChat AI 聊天
func (uc *ToolUsecase) AIChat(ctx context.Context, message string, model string, history []ChatMessage) (string, error) {
	// 简单模拟响应，便于测试
	response := fmt.Sprintf("AI 响应：\n\n你发送的消息是：「%s」\n\n模型：%s\n历史消息数：%d",
		message, model, len(history))
	return response, nil
}

// AIProcessText AI 文本处理
func (uc *ToolUsecase) AIProcessText(ctx context.Context, text, action, targetLang string) (string, error) {
	var result string

	switch action {
	case "translate":
		if targetLang == "" {
			targetLang = "中文"
		}
		result = fmt.Sprintf("翻译结果（%s）：\n原文：%s\n译文：%s", targetLang, text, text)
	case "summarize":
		result = fmt.Sprintf("总结：\n原文长度：%d字\n%s", len(text), text)
	case "polish":
		result = fmt.Sprintf("润色结果：\n%s", text)
	case "grammar":
		result = fmt.Sprintf("语法修正：\n%s", text)
	default:
		result = fmt.Sprintf("操作「%s」处理结果：\n%s", action, text)
	}

	return result, nil
}
