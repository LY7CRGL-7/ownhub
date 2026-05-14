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
	return &ToolService{uc: uc}
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
