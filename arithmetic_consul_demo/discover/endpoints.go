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
func MakeDiscoverEndpoint(ctx context.Context, client consul.Client, logger log.Logger) endpoint.Endpoint {
	serviceName := "arithmetic"
	tags := []string{"arithmetic", "raysonxin"}
	passingOnly := true
	duration := 500 * time.Millisecond

	//基于consul客户端、服务名称、服务标签等信息，
	// 创建consul的连接实例，
	// 可实时查询服务实例的状态信息
	instancer := consul.NewInstancer(client, logger, serviceName, tags, passingOnly)

	//针对calculate接口创建sd.Factory
	factory := arithmeticFactory(ctx, "POST", "calculate")

	//使用consul连接实例（发现服务系统）、factory创建sd.Factory
	endpointer := sd.NewEndpointer(instancer, factory, logger)

	//创建RoundRibbon负载均衡器
	balancer := lb.NewRoundRobin(endpointer)

	//为负载均衡器增加重试功能，同时该对象为endpoint.Endpoint
	retry := lb.Retry(1, duration, balancer)

	return retry
}
