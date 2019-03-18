package service

import (
	"context"
	"github.com/pkg/errors"
	"github.com/raysonxin/gokit-article-demo/arithmetic_grpc_demo/pb"
)

func EncodeGRPCArithmeticRequest(_ context.Context, r interface{}) (interface{}, error) {
	req := r.(ArithmeticRequest)
	return &pb.ArithmeticRequest{
		RequestType: req.RequestType,
		A:           int32(req.A),
		B:           int32(req.B),
	}, nil
}

func DecodeGRPCArithmeticRequest(ctx context.Context, r interface{}) (interface{}, error) {
	req := r.(*pb.ArithmeticRequest)
	return ArithmeticRequest{
		RequestType: req.RequestType,
		A:           int(req.A),
		B:           int(req.B),
	}, nil
}

func EncodeGRPCArithmeticResponse(_ context.Context, r interface{}) (interface{}, error) {
	resp := r.(ArithmeticResponse)

	if resp.Error != nil {
		return &pb.ArithmeticResponse{
			Result: int32(resp.Result),
			Err:    resp.Error.Error(),
		}, nil
	}

	return &pb.ArithmeticResponse{
		Result: int32(resp.Result),
		Err:    "",
	}, nil
}

func DecodeGRPCArithmeticResponse(_ context.Context, r interface{}) (interface{}, error) {
	resp := r.(*pb.ArithmeticResponse)
	return ArithmeticResponse{
		Result: int(resp.Result),
		Error:  errors.New(resp.Err),
	}, nil
}
