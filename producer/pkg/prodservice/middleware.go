package prodservice

import (
	"context"
	"time"

	"github.com/go-kit/log"
)

type Middleware func(Service) Service

type loggingMiddleware struct {
	log  log.Logger
	next Service
}

func LoggingMiddleware(log log.Logger) Middleware {
	return func(s Service) Service {
		return &loggingMiddleware{
			log:  log,
			next: s,
		}
	}
}

func (mw loggingMiddleware) ProduceMail(ctx context.Context, email, password string) (err error) {
	defer func(start time.Time) {
		mw.log.Log(
			"method",
			"ProduceMail",
			"email",
			email,
			"took",
			time.Since(start),
			"err",
			err,
		)
	}(time.Now())
	return mw.next.ProduceMail(ctx, email, password)
}

type instrumentingMiddleware struct {
	next Service
}

func InstrumentingMiddleware() Middleware {
	return func(s Service) Service {
		return &instrumentingMiddleware{
			next: s,
		}
	}
}

func (mw instrumentingMiddleware) ProduceMail(
	ctx context.Context,
	email, password string,
) (err error) {
	return mw.next.ProduceMail(ctx, email, password)
}
