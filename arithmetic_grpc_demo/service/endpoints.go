package service

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
)

// CalculateEndpoint define endpoint
type ArithmeticEndpoints struct {
	CalculateEndpoint   endpoint.Endpoint
	HealthCheckEndpoint endpoint.Endpoint
	AuthEndpoint        endpoint.Endpoint
}

// ArithmeticRequest define request struct
type ArithmeticRequest struct {
	RequestType string `json:"request_type"`
	A           int    `json:"a"`
	B           int    `json:"b"`
}

// ArithmeticResponse define response struct
type ArithmeticResponse struct {
	Result int   `json:"result"`
	Error  error `json:"error"`
}

// MakeArithmeticEndpoint make endpoint
func MakeArithmeticEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ArithmeticRequest)

		var (
			res, a, b int
			calError  error
		)

		a = req.A
		b = req.B

		res, calError = svc.Calculate(ctx, req.RequestType, a, b)

		return ArithmeticResponse{Result: res, Error: calError}, nil
	}
}

// HealthRequest 健康检查请求结构
type HealthRequest struct{}

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status bool `json:"status"`
}

// MakeHealthCheckEndpoint 创建健康检查Endpoint
func MakeHealthCheckEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthResponse{status}, nil
	}
}

// AuthRequest
type AuthRequest struct {
	Name string `json:"name"`
	Pwd  string `json:"pwd"`
}

// AuthResponse
type AuthResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Error   string `json:"error"`
}

func MakeAuthEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(AuthRequest)

		token, err := svc.Login(req.Name, req.Pwd)

		var resp AuthResponse
		if err != nil {
			resp = AuthResponse{
				Success: err == nil,
				Token:   token,
				Error:   err.Error(),
			}
		} else {
			resp = AuthResponse{
				Success: err == nil,
				Token:   token,
			}
		}

		return resp, nil
	}
}

func (ae ArithmeticEndpoints) Calculate(ctx context.Context, reqType string, a, b int) (res int, err error) {
	//ctx := context.Background()
	resp, err := ae.CalculateEndpoint(ctx, ArithmeticRequest{
		RequestType: reqType,
		A:           a,
		B:           b,
	})
	if err != nil {
		return 0, err
	}
	response := resp.(ArithmeticResponse)
	return response.Result, nil
}

func (ae ArithmeticEndpoints) HealthCheck() bool {
	return false
}

// HealthCheck
func (ae ArithmeticEndpoints) Login(name, pwd string) (string, error) {
	return "", errors.New("not implemented")
}
