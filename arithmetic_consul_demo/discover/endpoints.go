package main

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	"time"
)

// MakeDiscoverEndpoint 使用consul.Client创建服务发现Endpoint
// 为了方便这里默认了一些参数
func MakeDiscoverEndpoint(ctx context.Context, client consul.Client, factory sd.Factory, logger log.Logger) endpoint.Endpoint {
	serviceName := "arithmetic"
	tags := []string{"arithmetic", "raysonxin"}
	passingOnly := true
	duration := 500 * time.Millisecond
	var ep endpoint.Endpoint

	instancer := consul.NewInstancer(client, logger, serviceName, tags, passingOnly)
	endpointer := sd.NewEndpointer(instancer, factory, logger)
	balancer := lb.NewRoundRobin(endpointer)
	retry := lb.Retry(1, duration, balancer)
	ep = retry

	return ep
}
