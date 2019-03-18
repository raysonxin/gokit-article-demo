package main

import (
	"context"
	"flag"
	"fmt"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/raysonxin/gokit-article-demo/arithmetic_grpc_demo/pb"
	"github.com/raysonxin/gokit-article-demo/arithmetic_grpc_demo/service"
	"google.golang.org/grpc"
	"time"
)

func main() {
	var (
		grpcAddr = flag.String("addr", ":9001", "gRPC address")
	)
	flag.Parse()

	ctx := context.Background()

	conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(1*time.Second))
	if err != nil {
		fmt.Println("gRPC dial err:", err)
	}
	defer conn.Close()

	svr := NewClient(conn)
	result, err := svr.Calculate(ctx, "Add", 10, 2)
	if err != nil {
		fmt.Println("calculate error", err.Error())

	}

	fmt.Println("result=", result)
}

func NewClient(conn *grpc.ClientConn) service.Service {
	var ep = grpctransport.NewClient(conn,
		"pb.ArithmeticService",
		"Calculate",
		service.EncodeGRPCArithmeticRequest,
		service.DecodeGRPCArithmeticResponse,
		pb.ArithmeticResponse{},
	).Endpoint()

	arithmeticEp := service.ArithmeticEndpoints{
		CalculateEndpoint: ep,
	}
	return arithmeticEp
}
