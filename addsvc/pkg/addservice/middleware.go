package addservice

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

type Middleware func(Service) Service

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{logger, next}
	}
}

func InstrumentingMiddleware(ints, chars metrics.Counter) Middleware {
	return func(next Service) Service {
		return instrumentingMiddleware{
			ints:  ints,
			chars: chars,
			next:  next,
		}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) Sum(ctx context.Context, a, b int) (v int, err error) {
	defer func() {
		mw.logger.Log("method", "Sum", "a", a, "b", b, "v", v, "err", err)
	}()
	return mw.next.Sum(ctx, a, b)
}
func (mw loggingMiddleware) Concat(ctx context.Context, a, b string) (v string, err error) {
	defer func() {
		mw.logger.Log("method", "Concat", "a", a, "b", b, "v", v, "err", err)
	}()
	return mw.next.Concat(ctx, a, b)
}

type instrumentingMiddleware struct {
	ints  metrics.Counter
	chars metrics.Counter
	next  Service
}

func (mw instrumentingMiddleware) Sum(ctx context.Context, a, b int) (v int, err error) {
	v, err = mw.next.Sum(ctx, a, b)
	mw.ints.Add(float64(v))
	return v, err
}
func (mw instrumentingMiddleware) Concat(ctx context.Context, a, b string) (v string, err error) {
	v, err = mw.next.Concat(ctx, a, b)
	mw.ints.Add(float64(len(v)))
	return mw.next.Concat(ctx, a, b)
}
