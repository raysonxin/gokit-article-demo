

```

func DynamicLimitter(interval int, burst int) endpoint.Middleware {
	bucket := rate.NewLimiter(rate.Every(time.Second*time.Duration(interval)), burst)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if !bucket.Allow() {
				return nil, ErrLimitExceed
			}
			return next(ctx, request)
		}
	}
}
```

```
endpoint := MakeArithmeticEndpoint(svc)
endpoint = DynamicLimitter(1, 3)(endpoint)

testEndp := MakeArithmeticEndpoint(svc)
testEndp = DynamicLimitter(1, 2)(testEndp)

```