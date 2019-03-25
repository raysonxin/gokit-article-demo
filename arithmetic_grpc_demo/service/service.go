package service

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidRequestType = errors.New("RequestType has only four type: Add,Subtract,Multiply,Divide")
)

// Service Define a service interface
type Service interface {
	Calculate(ctx context.Context, reqType string, a, b int) (int, error)

	// HealthCheck check service health status
	HealthCheck() bool

	// HealthCheck
	Login(name, pwd string) (string, error)
}

//ArithmeticService implement Service interface
type ArithmeticService struct {
}

// Calculate 实现Service接口
func (s ArithmeticService) Calculate(ctx context.Context, reqType string, a, b int) (res int, err error) {

	if strings.EqualFold(reqType, "Add") {
		res = a + b
		return
	} else if strings.EqualFold(reqType, "Subtract") {
		res = a - b
		return
	} else if strings.EqualFold(reqType, "Multiply") {
		res = a * b
		return
	} else if strings.EqualFold(reqType, "Divide") {
		if b == 0 {
			res, err = 0, errors.New("the dividend can not be zero!")
			return
		}
		res, err = a/b, nil
	} else {
		res, err = 0, ErrInvalidRequestType
	}
	return
}

// HealthCheck implement Service method
// 用于检查服务的健康状态，这里仅仅返回true。
func (s ArithmeticService) HealthCheck() bool {
	return true
}

func (s ArithmeticService) Login(name, pwd string) (string, error) {
	if name == "name" && pwd == "pwd" {
		token, err := Sign(name, pwd)
		return token, err
	}

	return "", errors.New("Your name or password dismatch")
}

// ServiceMiddleware define service middleware
type ServiceMiddleware func(Service) Service
