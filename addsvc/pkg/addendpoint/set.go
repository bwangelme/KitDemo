package addendpoint

import (
	"context"
	"kitdemo/addsvc/pkg/addservice"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	"github.com/go-kit/kit/endpoint"
)

type Set struct {
	SumEndpoint    endpoint.Endpoint
	ConcatEndpoint endpoint.Endpoint
}

func New(svc addservice.Service, logger log.Logger, duration metrics.Histogram) Set {
	var sumEndpoint endpoint.Endpoint
	{
		sumEndpoint = MakeSumEndpoint(svc)
		//TODO: zipkin trace
		//TODO: open trace
		//TODO: 频率限制和熔断
		sumEndpoint = LoggingMiddleware(log.With(logger, "method", "Sum"))(sumEndpoint)
		sumEndpoint = InstrumentingMiddleware(duration.With("method", "Sum"))(sumEndpoint)
	}

	var concatEndpoint endpoint.Endpoint
	{
		concatEndpoint = MakeConcatEndpoint(svc)
		//TODO: zipkin trace
		//TODO: open trace
		//TODO: 频率限制和熔断
		concatEndpoint = LoggingMiddleware(log.With(logger, "method", "Concat"))(concatEndpoint)
		concatEndpoint = InstrumentingMiddleware(duration.With("method", "Concat"))(concatEndpoint)
	}

	return Set{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}
}

//Sum Set 实现 Service 接口
func (s Set) Sum(ctx context.Context, a, b int) (int, error) {
	resp, err := s.SumEndpoint(ctx, SumRequest{A: a, B: b})
	if err != nil {
		return 0, err
	}
	response := resp.(SumResponse)
	return response.V, response.Err
}

//Concat Set 实现 Service 接口
func (s Set) Concat(ctx context.Context, a, b string) (string, error) {
	resp, err := s.ConcatEndpoint(ctx, ConcatRequest{A: a, B: b})
	if err != nil {
		return "", err
	}
	response := resp.(ConcatResponse)
	return response.V, response.Err
}

func MakeSumEndpoint(s addservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(SumRequest)
		v, err := s.Sum(ctx, req.A, req.B)
		return SumResponse{V: v, Err: err}, nil
	}
}

func MakeConcatEndpoint(s addservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ConcatRequest)
		v, err := s.Concat(ctx, req.A, req.B)
		return ConcatResponse{V: v, Err: err}, nil
	}
}

// 类型断言，保证 SumResponse 和 ConcatResponse 都实现了 Failer 接口
var (
	_ endpoint.Failer = SumResponse{}
	_ endpoint.Failer = ConcatResponse{}
)

type SumRequest struct {
	A, B int
}

type SumResponse struct {
	V   int   `json:"v"`
	Err error `json:"-"` // 应该被 Failed/errorEncoder 拦截住
}

//Failed 实现了 endpoint.Failer 接口
func (r SumResponse) Failed() error { return r.Err }

//ConcatRequest Concat RPC 接口的请求参数
type ConcatRequest struct {
	A, B string
}

//ConcatResponse 返回 Concat 的响应值
type ConcatResponse struct {
	V   string `json:"v"`
	Err error  `json:"-"`
}

func (r ConcatResponse) Failed() error {
	return r.Err
}
