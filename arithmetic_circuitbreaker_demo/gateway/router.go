package main

import (
	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/log"
	"net/http"
	"strings"
	"sync"
)

type HystrixRouter struct {
	next        http.Handler
	svcMap      *sync.Map
	logger      log.Logger
	fallbackMsg string
}

func (router *HystrixRouter) Routes(next http.Handler, fbMsg string, logger log.Logger) http.Handler {
	return HystrixRouter{
		next:        next,
		svcMap:      &sync.Map{},
		logger:      logger,
		fallbackMsg: fbMsg,
	}
}

func (router HystrixRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//查询原始请求路径，如：/arithmetic/calculate/10/5
	reqPath := r.URL.Path
	if reqPath == "" {
		return
	}
	//按照分隔符'/'对路径进行分解，获取服务名称serviceName
	pathArray := strings.Split(reqPath, "/")
	serviceName := pathArray[1]
	if _, ok := router.svcMap.Load(serviceName); !ok {

		hystrix.ConfigureCommand(serviceName, hystrix.CommandConfig{Timeout: 1000})

		router.svcMap.Store(serviceName, serviceName)
	}

	var resp interface{}

	if err := hystrix.Do(serviceName, func() (err error) {
		router.next.ServeHTTP(w, r)
		return nil
	}, func(err error) error {
		router.logger.Log("fallback error description", err.Error())

		resp = "Service unavailable"
		return nil
	}); err != nil {
		resp = "Hystrix Do error" + err.Error()
	}

}
