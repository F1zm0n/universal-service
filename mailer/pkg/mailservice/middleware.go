package mailservice

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/google/uuid"
)

type Middleware func(Service) Service

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) VerifyMail(ctx context.Context, id uuid.UUID) (err error) {
	defer func(start time.Time) {
		mw.logger.Log("method", "VerifyMail", "id", id, "took", time.Since(start), "err", err)
	}(time.Now())
	return mw.next.VerifyMail(ctx, id)
}

func (mw loggingMiddleware) SendEmail(ctx context.Context, ver VerDto) (err error) {
	defer func(start time.Time) {
		mw.logger.Log(
			"method",
			"SendEmail",
			"email",
			ver.Email,
			"took",
			time.Since(start),
			"err",
			err,
		)
	}(time.Now())
	return mw.next.SendEmail(ctx, ver)
}

func LoggingMiddleware(log log.Logger) Middleware {
	return func(s Service) Service {
		return &loggingMiddleware{
			next:   s,
			logger: log,
		}
	}
}

type instrumentingMiddleware struct {
	next Service
}

func (mw instrumentingMiddleware) VerifyMail(ctx context.Context, id uuid.UUID) (err error) {
	return mw.next.VerifyMail(ctx, id)
}

func (mw instrumentingMiddleware) SendEmail(ctx context.Context, ver VerDto) (err error) {
	return mw.next.SendEmail(ctx, ver)
}

func InstrumentingMiddleware() Middleware {
	return func(s Service) Service {
		return &instrumentingMiddleware{
			next: s,
		}
	}
}
