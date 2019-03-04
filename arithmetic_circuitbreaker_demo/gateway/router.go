package main

import (
	"errors"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
	"github.com/openzipkin/zipkin-go"
	zipkinhttpsvr "github.com/openzipkin/zipkin-go/middleware/http"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
)

type HystrixRouter struct {
	next         http.Handler
	svcMap       *sync.Map
	logger       log.Logger
	fallbackMsg  string
	consulClient *api.Client
	tracer       *zipkin.Tracer
}

func Routes(client *api.Client, zikkinTracer *zipkin.Tracer, fbMsg string, logger log.Logger) http.Handler {
	return HystrixRouter{
		//next:         next,
		svcMap:       &sync.Map{},
		logger:       logger,
		fallbackMsg:  fbMsg,
		consulClient: client,
		tracer:       zikkinTracer,
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

	err := hystrix.Do(serviceName, func() (err error) {

		//调用consul api查询serviceName的服务实例列表
		result, _, err := router.consulClient.Catalog().Service(serviceName, "", nil)
		if err != nil {
			router.logger.Log("ReverseProxy failed", "query service instace error", err.Error())
			return
		}

		if len(result) == 0 {
			router.logger.Log("ReverseProxy failed", "no such service instance", serviceName)
			return errors.New("no such service instance")
		}

		director := func(req *http.Request) {
			//重新组织请求路径，去掉服务名称部分
			destPath := strings.Join(pathArray[2:], "/")

			//随机选择一个服务实例
			tgt := result[rand.Int()%len(result)]
			router.logger.Log("service id", tgt.ServiceID)

			//设置代理服务地址信息
			req.URL.Scheme = "http"
			req.URL.Host = fmt.Sprintf("%s:%d", tgt.ServiceAddress, tgt.ServicePort)
			req.URL.Path = "/" + destPath
		}

		var proxyError error = nil
		// 为反向代理增加追踪逻辑，使用如下RoundTrip代替默认Transport
		roundTrip, _ := zipkinhttpsvr.NewTransport(router.tracer, zipkinhttpsvr.TransportTrace(true))

		errorHandler := func(ew http.ResponseWriter, er *http.Request, err error) {
			proxyError = err
		}

		proxy := &httputil.ReverseProxy{
			Director:     director,
			Transport:    roundTrip,
			ErrorHandler: errorHandler,
		}
		proxy.ServeHTTP(w, r)

		return proxyError
	}, func(err error) error {
		router.logger.Log("fallback error description", err.Error())

		return errors.New(router.fallbackMsg)
	})

	if err != nil {
		//resp.Err = "Hystrix Do error" + err.Error()
		//datas, _ := json.Marshal(resp)
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
}
