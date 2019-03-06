package main

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	ctx := context.Background()
	errChan := make(chan error)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var svc Service
	svc = ArithmeticService{}

	// add logging middleware
	svc = LoggingMiddleware(logger)(svc)

	endpoint := MakeArithmeticEndpoint(svc)

	// add ratelimit,refill every second,set capacity 3
	//ratebucket := ratelimit.NewBucket(time.Second*1, 3)
	//endpoint = NewTokenBucketLimitterWithJuju(ratebucket)(endpoint)

	//add ratelimit,refill every second,set capacity 3
	//	ratebucket := rate.NewLimiter(rate.Every(time.Second*1), 3)
	//endpoint = NewTokenBucketLimitterWithBuildIn(ratebucket)(endpoint)
	endpoint = DynamicLimitter(1, 3)(endpoint)

	testEndp := MakeArithmeticEndpoint(svc)
	testEndp = DynamicLimitter(1, 2)(testEndp)

	arithmeticEndp := ArithmeticEndpoint{
		CalculateEndpoint: endpoint,
		TestEndpoint:      testEndp,
	}

	r := MakeHttpHandler(ctx, arithmeticEndp, logger)

	go func() {
		fmt.Println("Http Server start at port:9000")
		handler := r
		errChan <- http.ListenAndServe(":9000", handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	fmt.Println(<-errChan)
}
