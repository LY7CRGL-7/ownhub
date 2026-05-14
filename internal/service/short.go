package service

import (
	"context"

	v1 "ownhub/api/ownhub/v1"
	"ownhub/internal/biz"
)

type ShortService struct {
	v1.UnimplementedShortServiceServer

	uc *biz.ShortUsecase
}

func NewShortService(uc *biz.ShortUsecase) *ShortService {
	return &ShortService{uc: uc}
}

func (s *ShortService) CreateShort(ctx context.Context, req *v1.CreateShortReq) (*v1.CreateShortReply, error) {
	short, shortURL, err := s.uc.Create(ctx, req.GetUrl(), req.GetSlug(), req.GetDescription())
	if err != nil {
		return nil, err
	}
	return &v1.CreateShortReply{
		Code:     short.Code,
		ShortUrl: shortURL,
	}, nil
}

func (s *ShortService) Redirect(ctx context.Context, req *v1.RedirectReq) (*v1.RedirectReply, error) {
	short, _, err := s.uc.Get(ctx, req.GetCode())
	if err != nil {
		return nil, err
	}
	return &v1.RedirectReply{Url: short.URL}, nil
}

func (s *ShortService) GetShort(ctx context.Context, req *v1.GetShortReq) (*v1.GetShortReply, error) {
	short, shortURL, err := s.uc.Get(ctx, req.GetCode())
	if err != nil {
		return nil, err
	}
	return &v1.GetShortReply{
		Code:          short.Code,
		Url:           short.URL,
		ShortUrl:      shortURL,
		Description:   short.Description,
		CreatedAtUnix: short.CreatedAt.Unix(),
	}, nil
}
