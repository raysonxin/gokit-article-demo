package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	var (
		consulHost  = flag.String("consul.host", "localhost", "consul ip address")
		consulPort  = flag.String("consul.port", "8500", "consul port")
		serviceHost = flag.String("service.host", "192.168.10.113", "service ip address")
		servicePort = flag.String("service.port", "9000", "service port")
		zipkinURL   = flag.String("zipkin.url", "http://localhost:9411/api/v2/spans", "Zipkin server url")
	)

	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	fieldKeys := []string{"method"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "raysonxin",
		Subsystem: "arithmetic_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)

	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "raysonxin",
		Subsystem: "arithemetic_service",
		Name:      "request_latency",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = *serviceHost + ":" + *servicePort
			serviceName   = "arithmetic-service"
			useNoopTracer = (*zipkinURL == "")
			reporter      = zipkinhttp.NewReporter(*zipkinURL)
		)
		defer reporter.Close()
		zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(
			reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		if !useNoopTracer {
			logger.Log("tracer", "Zipkin", "type", "Native", "URL", *zipkinURL)
		}
	}

	//add ratelimit,refill every second,set capacity 3
	ratebucket := rate.NewLimiter(rate.Every(time.Second*1), 100)

	var svc Service
	svc = ArithmeticService{}

	// add logging middleware to service
	svc = LoggingMiddleware(logger)(svc)
	svc = Metrics(requestCount, requestLatency)(svc)

	calEndpoint := MakeArithmeticEndpoint(svc)
	calEndpoint = NewTokenBucketLimitterWithBuildIn(ratebucket)(calEndpoint)
	calEndpoint = kitzipkin.TraceEndpoint(zipkinTracer, "calculate-endpoint")(calEndpoint)
	calEndpoint = kitjwt.NewParser(jwtKeyFunc, jwt.SigningMethodHS256, kitjwt.StandardClaimsFactory)(calEndpoint)

	//创建健康检查的Endpoint
	healthEndpoint := MakeHealthCheckEndpoint(svc)
	healthEndpoint = NewTokenBucketLimitterWithBuildIn(ratebucket)(healthEndpoint)
	healthEndpoint = kitzipkin.TraceEndpoint(zipkinTracer, "health-endpoint")(healthEndpoint)

	//身份认证Endpoint
	authEndpoint := MakeAuthEndpoint(svc)
	authEndpoint = NewTokenBucketLimitterWithBuildIn(ratebucket)(authEndpoint)
	authEndpoint = kitzipkin.TraceEndpoint(zipkinTracer, "login-endpoint")(authEndpoint)

	//把算术运算Endpoint\健康检查、登录Endpoint封装至ArithmeticEndpoints
	endpts := ArithmeticEndpoints{
		ArithmeticEndpoint:  calEndpoint,
		HealthCheckEndpoint: healthEndpoint,
		AuthEndpoint:        authEndpoint,
	}

	//创建http.Handler
	r := MakeHttpHandler(ctx, endpts, zipkinTracer, logger)

	//创建注册对象
	registar := Register(*consulHost, *consulPort, *serviceHost, *servicePort, logger)

	go func() {
		fmt.Println("Http Server start at port:" + *servicePort)
		//启动前执行注册
		registar.Register()
		handler := r
		errChan <- http.ListenAndServe(":"+*servicePort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	//服务退出取消注册
	registar.Deregister()
	fmt.Println(error)
}
