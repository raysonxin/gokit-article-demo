目前，几乎所有的软件系统都具备日志功能，通过日志我们可以在软件运行异常时定位软件遇到的问题，还原应用程序异常时的运行状态。

虽然系统上线前经过了严格的测试工作，但是生产环境业务的复杂性、不可预测性使得软件工程师无法确保系统上线后不会发生故障。为了能够在系统发生异常时对系统故障进行分析与定位，引入日志系统成为所有软件系统研发的必然选择。

# Gokit日志记录

在上篇文章[《Gokit微服务-REST HTTP服务》](https://mp.weixin.qq.com/s?__biz=MzI0NTE4NDg0NA==&mid=2247483658&idx=1&sn=68f96f528d2aca5d3b22124e7fc97906&chksm=e9532129de24a83fa616cf16075c614099cd07ebcfd8bb73dd2ddb3011cef1f76f34b0190015&mpshare=1&scene=1&srcid=#rd)中实现了算术运算的HTTP服务，这篇文章将基于Gokit中间件机制为其增加日志记录功能。

> 本质上讲，Gokit中间件采用了装饰者模式，传入Endpoint对象，封装部分业务逻辑，然后返回Endpoint对象。

## Step-1：创建Middleware

打开`service.go`文件，加入如下代码：
```
// ServiceMiddleware define service middleware
type ServiceMiddleware func(Service) Service
```
## Step-2：创建日志中间件

新建文件`loggings.go`，新建类型`loggingMiddleware`，该类型中嵌入了`Service`，还包含一个`logger`属性，代码如下所示：

```
// loggingMiddleware Make a new type
// that contains Service interface and logger instance
type loggingMiddleware struct {
	Service
	logger log.Logger
}
```

下来创建一个方法`LoggingMiddleware`把日志记录对象嵌入中间件。该方法接受日志对象，返回`ServiceMiddleware`，而`ServiceMiddleware`可以传入`Service`对象，这样就可以对`Service`增加一层装饰。代码如下：

```
// LoggingMiddleware make logging middleware
func LoggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next Service) Service {
		return loggingMiddleware{next, logger}
	}
}
```

接下来就可以让新的类型`loggingMiddleware`实现`Service`的接口方法了。实现方法时可以在其中使用日志对象记录调用方法、调用时间、传入参数、输出结果、调用耗时等信息。下面以`Add`方法为例进行实现，其他方法与之类似：
```
func (mw loggingMiddleware) Add(a, b int) (ret int) {

	defer func(beign time.Time) {
		mw.logger.Log(
			"function", "Add",
			"a", a,
			"b", b,
			"result", ret,
			"took", time.Since(beign),
		)
	}(time.Now())

	ret = mw.Service.Add(a, b)
	return ret
}
```

## Step-3：修改main方法

打开`main.go`，调用`LoggingMiddleware`创建日志中间件实现对`svc`的包装，代码如下所示（带有注释的一行即为新增代码）：

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

## Step-4：编译&运行

在控制台编译并运行应用程序，然后通过Postman请求接口进行测试，即可看到输出的日志信息：
```
ts=2019-02-18T05:43:57.902971Z caller=logging.go:25 function=Add a=10 b=1 result=11 took=0s
ts=2019-02-18T05:44:10.116234Z caller=logging.go:25 function=Add a=10 b=1 result=11 took=0s
ts=2019-02-18T05:44:11.2682718Z caller=logging.go:25 function=Add a=10 b=1 result=11 took=0s
```

# 总结

本文借助Gokit的中间件机制为微服务增加了日志功能。由于Gokit中间件采用装饰者模式，新增的日志功能对Endpoint、Service、Transport三个层次均无代码入侵，实现即插即用的效果，这一机制在开发中将非常有利于团队之间的配合。当然，本文采用的日志记录仅仅是通过控制台输出，还无法真正应用于生产环境，今后有时间继续研究其他方式。

本文代码可通过[github](https://github.com/raysonxin/gokit-article-demo)获取。