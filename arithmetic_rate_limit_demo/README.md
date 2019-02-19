由于业务应用系统的负载能力有限，为了防止非预期的请求对系统压力过大而拖垮业务应用系统，每个API接口都是有访问上限的。API接口的流量控制策略：分流、降级、限流等。本文讨论限流策略，虽然降低了服务接口的访问频率和并发量，却换取服务接口和业务应用系统的高可用。

> 限流的目的是通过对并发访问/请求进行限速，或者对一个时间窗口内的请求进行限速来保护系统，一旦达到限制速率则可以拒绝服务、排队或等待、降级等处理。

# 限流算法

常用的限流算法有两种：漏桶算法和令牌桶算法。

## 漏桶算法

漏桶算法(Leaky Bucket)是网络世界中流量整形（Traffic Shaping）或速率限制（Rate Limiting）时经常使用的一种算法，它的主要目的是控制数据注入到网络的速率，平滑网络上的突发流量。漏桶算法提供了一种机制，通过它，突发流量可以被整形以便为网络提供一个稳定的流量。

漏桶算法思路很简单，水（请求）先进入到漏桶里，漏桶以一定的速度出水（接口有响应速率），当水流入速度过大会直接溢出（访问频率超过接口响应速率），然后就拒绝请求，可以看出漏桶算法能强行限制数据的传输速率。示意图如下：

![](https://mmbiz.qpic.cn/mmbiz_png/mZx0iasykfluUNN9XcYGLEKpq98pAOCGuiaUOfiaNq0jvWCho3dhF904ibIKcCHM9Cx4Z5d3CD9A5nqgMeP64ZLqjw/0?wx_fmt=png)

因为漏桶的漏出速率是固定的参数,所以即使网络中不存在资源冲突（没有发生拥塞），漏桶算法也不能使流突发（burst）到端口速率。因此，漏桶算法对于存在突发特性的流量来说缺乏效率。

## 令牌桶算法

令牌桶算法是网络流量整形（Traffic Shaping）和速率限制（Rate Limiting）中最常使用的一种算法。典型情况下，令牌桶算法用来控制发送到网络上的数据的数目，并允许突发数据的发送。

令牌桶算法的原理是系统会以一个恒定的速度往桶里放入令牌，而如果请求需要被处理，则需要先从桶里获取一个令牌，当桶里没有令牌可取时，则拒绝服务。从原理上看，令牌桶算法和漏桶算法是相反的，一个“进水”，一个是“漏水”。

![](https://mmbiz.qpic.cn/mmbiz_jpg/mZx0iasykfluUNN9XcYGLEKpq98pAOCGuVxUHqsD0hafFkeF9UMBAV2FZgvqbLL8mjJqlmmgUa9yY89mr7IucHg/0?wx_fmt=jpeg)

令牌桶的另外一个好处是可以方便的改变速度。 一旦需要提高速率，则按需提高放入桶中的令牌的速率。 一般会定时（比如100毫秒）往桶中增加一定数量的令牌，有些变种算法则实时的计算应该增加的令牌的数量。

# Gokit微服务限流实现

结合以上分析我将基于Gokit实现微服务的限流功能。通过查阅`gokit/kit/ratelimit`源码，发现gokit基于go包`golang.org/x/time/rate`内置了一种实现；另外，在此之前gokit默认使用的`juju/ratelimit`实现方案（目前官方已经移除），我将基于两种方式分别进行实现。

与之前两篇文章不同，本次实现将基于gokit内建的类型`endpoint.Middleware`，该类型实际上是一个function，使用装饰者模式实现对Endpoint的封装。定义如下：

```
# Go-kit Middleware Endpoint
type Middleware func(Endpoint) Endpoint
```

## juju/ratelimit方案

本文示例将继续在上篇文章代码基础上进行完善（地址附文末），前两篇忘记放地址。

### Step-1：创建限流器

首先，使用如下命令安装最新版本的`juju/ratelimit`库：

```
go get github.com/juju/ratelimit
```

然后，新建go文件命名为`instrument.go`，实现限流方法：参数为令牌桶（bkt）返回`endpoint.Middleware`。使用令牌桶的`TakeAvaiable`方法获取令牌，若获取成功则继续执行，若获取失败则返回异常（即限流）。代码如下：

```
var ErrLimitExceed = errors.New("Rate limit exceed!")

// NewTokenBucketLimitterWithJuju 使用juju/ratelimit创建限流中间件
func NewTokenBucketLimitterWithJuju(bkt *ratelimit.Bucket) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if bkt.TakeAvailable(1) == 0 {
				return nil, ErrLimitExceed
			}
			return next(ctx, request)
		}
	}
}
```
### Step-2：修改main

下来就是使用`juju/ratelimit`创建令牌桶（每秒刷新一次，容量为3），然后调用`Step-1`实现限流方法对Endpoint进行装饰。在main方法中增加如下代码。

```
// add ratelimit,refill every second,set capacity 3
ratebucket := ratelimit.NewBucket(time.Second*1, 3)
endpoint = NewTokenBucketLimitterWithJuju(ratebucket)(endpoint)
```

修改后，完整代码如下：

```
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
	ratebucket := ratelimit.NewBucket(time.Second*1, 3)
	endpoint = NewTokenBucketLimitterWithJuju(ratebucket)(endpoint)

	r := MakeHttpHandler(ctx, endpoint, logger)

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
```

### Step-3：编译&运行

在控制台编译并运行应用程序，然后通过Postman请求接口进行测试，即可看到输出的日志信息：
```
ts=2019-02-19T03:20:13.1908613Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T03:20:13.7144627Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T03:20:14.2276079Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T03:20:14.7414288Z caller=server.go:112 err="Rate limit exceed!"
ts=2019-02-19T03:20:15.2091773Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T03:20:16.0261559Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T03:20:16.6406654Z caller=server.go:112 err="Rate limit exceed!"
ts=2019-02-19T03:20:17.1912533Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T03:20:17.7828906Z caller=server.go:112 err="Rate limit exceed!"
```

从日志中可以看到，请求中出现了`Rate limit exceed!`，即限流器把令牌发完了将请求中断，服务不可用；接下来继续访问时，服务恢复，即限流器恢复填满令牌桶。

## gokit内置实现方案

### Step-1：创建限流器

首先下载依赖的go`/time/rate`包，安装方式如下（无法直接使用go get指令）：
```
git clone https://github.com/golang/time.git [Your GOPATH]/src/golang.org/x
```
然后在`instrument.go`中添加方法`NewTokenBucketLimitterWithBuildIn`，在其中使用`x/time/rate`实现限流方法：

```
// NewTokenBucketLimitterWithBuildIn 使用x/time/rate创建限流中间件
func NewTokenBucketLimitterWithBuildIn(bkt *rate.Limiter) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if !bkt.Allow() {
				return nil, ErrLimitExceed
			}
			return next(ctx, request)
		}
	}
}
```

### Step-2：修改main

将限流方法封装改为如下实现：

```	
//add ratelimit,refill every second,set capacity 3
ratebucket := rate.NewLimiter(rate.Every(time.Second*1), 3)
endpoint = NewTokenBucketLimitterWithBuildIn(ratebucket)(endpoint)
```

### Step-3：编译&运行

在控制台编译并运行应用程序，然后通过Postman请求接口进行测试，即可看到输出的日志信息：

```
ts=2019-02-19T06:03:26.8650217Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T06:03:27.5747177Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T06:03:28.1274404Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T06:03:28.5892068Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T06:03:29.1327522Z caller=server.go:112 err="Rate limit exceed!"
ts=2019-02-19T06:03:29.59453Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T06:03:30.2138805Z caller=server.go:112 err="Rate limit exceed!"
ts=2019-02-19T06:03:30.6257682Z caller=logging.go:41 function=Subtract a=10 b=1 result=9 took=0s
ts=2019-02-19T06:03:31.2772011Z caller=server.go:112 err="Rate limit exceed!"
```

由日志可以看出效果与`juju/ratelimit`方案一样。

# 总结

本文首先介绍了两种常用的限流算法漏桶算法和令牌桶算法，然后通过两种方案（`juju/ratelimit`和gokit内置库）实现服务限流。

服务开发过程中我们需要充分考虑服务的可用性，尤其是那些比较消耗系统资源的服务，为其增加限流机制，确保服务稳定可靠运行。


> 图片来自互联网。

- 本文代码地址：https://github.com/raysonxin/gokit-article-demo
- Token Bucket：https://en.wikipedia.org/wiki/Token_bucket
- juju/ratelimit：https://github.com/juju/ratelimit

