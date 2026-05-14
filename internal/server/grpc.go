package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	v1 "ownhub/api/ownhub/v1"
	"ownhub/internal/conf"
	"ownhub/internal/service"
)

func NewGRPCServer(c *conf.Server, shortSvc *service.ShortService, toolSvc *service.ToolService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(recovery.Recovery()),
		grpc.Logger(logger),
	}
	if c != nil && c.Grpc != nil {
		opts = append(opts, grpc.Network(c.Grpc.Network))
		opts = append(opts, grpc.Address(c.Grpc.Addr))
		if c.Grpc.Timeout != nil {
			opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
		}
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterShortServiceServer(srv, shortSvc)
	v1.RegisterToolServiceServer(srv, toolSvc)
	return srv
}
