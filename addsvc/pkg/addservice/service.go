package addservice

import (
	"context"
	"errors"

	"github.com/go-kit/kit/log"

	"github.com/go-kit/kit/metrics"
)

type Service interface {
	Sum(ctx context.Context, a, b int) (int, error)
	Concat(ctx context.Context, a, b string) (string, error)
}

//New 返回一个基础的 addsvc.Service 服务，安装了统计和日志的中间件
func New(logger log.Logger, ints, chars metrics.Counter) Service {
	var svc Service
	{
		svc = NewBasicService()
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(ints, chars)(svc)
	}
	return svc
}

var (
	//ErrZeroPara 假设的业务逻辑，参数不能是0
	ErrZeroPara = errors.New("Number cannot be zeore")

	//ErrIntOverflow
	ErrIntOverflow = errors.New("integer overflow")

	//ErrMaxSizeExceeded
	ErrMaxSizeExceeded = errors.New("result exceeds maximum size")
)

const (
	intMax = 1<<31 - 1
	intMin = -(intMax + 1)
	maxLen = 10
)

type basicService struct{}

func NewBasicService() Service {
	return basicService{}
}

func (s basicService) Sum(_ context.Context, a, b int) (int, error) {
	if a == 0 || b == 0 {
		return 0, ErrZeroPara
	}
	if (b > 0 && a > (intMax-b)) || (b < 0 && a < (intMin-b)) {
		return 0, ErrIntOverflow
	}
	return a + b, nil
}

func (s basicService) Concat(_ context.Context, a, b string) (string, error) {
	if len(a)+len(b) > maxLen {
		return "", ErrMaxSizeExceeded
	}
	return a + b, nil
}
