package prodservice

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/google/uuid"
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

func (mw loggingMiddleware) ProduceVer(
	ctx context.Context,
	verId uuid.UUID,
) (err error) {
	defer func(start time.Time) {
		mw.log.Log(
			"method",
			"ProduceVer",
			"ver_id",
			verId,
			"took",
			time.Since(start),
			"err",
			err,
		)
	}(time.Now())
	return mw.next.ProduceVer(ctx, verId)
}

func (mw loggingMiddleware) ProduceRegister(
	ctx context.Context,
	email, password string,
) (err error) {
	defer func(start time.Time) {
		mw.log.Log(
			"method",
			"ProduceRegister",
			"email",
			email,
			"took",
			time.Since(start),
			"err",
			err,
		)
	}(time.Now())
	return mw.next.ProduceRegister(ctx, email, password)
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

func (mw instrumentingMiddleware) ProduceRegister(
	ctx context.Context,
	email, password string,
) (err error) {
	return mw.next.ProduceRegister(ctx, email, password)
}

func (mw instrumentingMiddleware) ProduceVer(
	ctx context.Context,
	verId uuid.UUID,
) (err error) {
	return mw.next.ProduceVer(ctx, verId)
}
