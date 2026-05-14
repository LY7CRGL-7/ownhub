//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"ownhub/internal/biz"
	"ownhub/internal/conf"
	"ownhub/internal/data"
	"ownhub/internal/server"
	"ownhub/internal/service"
)

// 👇 把 Bootstrap 拆成各个子配置
func newServer(c *conf.Bootstrap) *conf.Server {
	return c.Server
}

func newData(c *conf.Bootstrap) *conf.Data {
	return c.Data
}

func newAppConf(c *conf.Bootstrap) *conf.App {
	return c.App
}

// 👇 组装 Provider
var ProviderSet = wire.NewSet(
	server.ProviderSet,
	data.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,

	newServer,
	newData,
	newAppConf,

	newApp,
)

// 👇 入口
func wireApp(*conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(ProviderSet))
}
