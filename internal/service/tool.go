package service

import (
	"context"
	"time"

	v1 "ownhub/api/ownhub/v1"
	"ownhub/internal/biz"
)

type ToolService struct {
	v1.UnimplementedToolServiceServer

	uc *biz.ToolUsecase
}

func NewToolService(uc *biz.ToolUsecase) *ToolService {
	return &ToolService{
		uc: uc,
	}
}

func (s *ToolService) Ping(ctx context.Context, _ *v1.PingReq) (*v1.PingReply, error) {
	serviceName, now := s.uc.Ping(ctx)
	return &v1.PingReply{
		Service: serviceName,
		Version: "dev",
		Now:     now,
	}, nil
}

func (s *ToolService) GetTimestamp(ctx context.Context, req *v1.GetTimestampReq) (*v1.GetTimestampReply, error) {
	now, loc, err := s.uc.Timestamp(ctx, req.GetTimezone())
	if err != nil {
		return nil, err
	}
	return &v1.GetTimestampReply{
		Unix:     now.Unix(),
		Rfc3339:  now.Format(time.RFC3339),
		Timezone: loc.String(),
	}, nil
}

func (s *ToolService) GenerateToken(ctx context.Context, req *v1.GenerateTokenReq) (*v1.GenerateTokenReply, error) {
	token, length, err := s.uc.GenerateToken(ctx, req.GetLength())
	if err != nil {
		return nil, err
	}
	return &v1.GenerateTokenReply{
		Token:  token,
		Length: length,
	}, nil
}

func (s *ToolService) DigestText(ctx context.Context, req *v1.DigestTextReq) (*v1.DigestTextReply, error) {
	algorithm, digest, err := s.uc.DigestText(ctx, req.GetText(), req.GetAlgorithm())
	if err != nil {
		return nil, err
	}
	return &v1.DigestTextReply{
		Algorithm: algorithm,
		Digest:    digest,
	}, nil
}

// AIChat AI 聊天
func (s *ToolService) AIChat(ctx context.Context, req *v1.AIChatReq) (*v1.AIChatReply, error) {
	// 转换历史消息
	var history []biz.ChatMessage
	for _, msg := range req.History {
		history = append(history, biz.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	response, err := s.uc.AIChat(ctx, req.Message, req.Model, history)
	if err != nil {
		return nil, err
	}

	return &v1.AIChatReply{
		Response: response,
		Model:    req.Model,
	}, nil
}

// AIProcessText AI 文本处理
func (s *ToolService) AIProcessText(ctx context.Context, req *v1.AIProcessTextReq) (*v1.AIProcessTextReply, error) {
	processed, err := s.uc.AIProcessText(ctx, req.Text, req.Action, req.TargetLang)
	if err != nil {
		return nil, err
	}

	return &v1.AIProcessTextReply{
		Original:  req.Text,
		Processed: processed,
		Action:    req.Action,
	}, nil
}
