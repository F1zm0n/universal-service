package authservice

import (
	"context"
	"time"

	"github.com/go-kit/log"
)

type Middleware func(svc Service) Service

type loggingMiddleware struct {
	next Service
	log  log.Logger
}
type instrumentingMiddleware struct {
	next Service
}

func (mw loggingMiddleware) Register(ctx context.Context, user User) (err error) {
	defer func(start time.Time) {
		mw.log.Log(
			"operation", "registering user",
			"email", user.Email,
			"error", err,
			"took", time.Since(start),
		)
	}(time.Now())
	return mw.next.Register(ctx, user)
}

func (mw loggingMiddleware) Login(ctx context.Context, user User) (token string, err error) {
	defer func(start time.Time) {
		mw.log.Log(
			"operation", "logging user",
			"email", user.Email,
			"error", err,
			"took", time.Since(start),
		)
	}(time.Now())
	return mw.next.Login(ctx, user)
}

func (mw instrumentingMiddleware) Register(ctx context.Context, user User) (err error) {
	return mw.next.Register(ctx, user)
}

func (mw instrumentingMiddleware) Login(ctx context.Context, user User) (token string, err error) {
	return mw.next.Login(ctx, user)
}

func LoggingMiddleware(l log.Logger) Middleware {
	return func(svc Service) Service {
		return &loggingMiddleware{
			next: svc,
			log:  l,
		}
	}
}

func InstrumentingMiddleware() Middleware {
	return func(svc Service) Service {
		return &instrumentingMiddleware{
			next: svc,
		}
	}
}
