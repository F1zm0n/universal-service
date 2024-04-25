package mailservice

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Middleware func(Service) Service

type loggingMiddleware struct {
	logger *slog.Logger
	next   Service
}

func (mw loggingMiddleware) VerifyMail(ctx context.Context, id uuid.UUID) (err error) {
	defer func(start time.Time) {
		mw.logger.Info(
			"verifying mail",
			slog.String("id", id.String()),
			slog.Duration("took", time.Since(start)),
			slog.Any("error", err),
		)
	}(time.Now())
	return mw.next.VerifyMail(ctx, id)
}

func (mw loggingMiddleware) SendEmail(ctx context.Context, ver VerDto) (err error) {
	defer func(start time.Time) {
		mw.logger.Info(
			"sending email",
			slog.String("email", ver.Email),
			slog.Duration("took", time.Since(start)),
			slog.Any("error", err),
		)
	}(time.Now())
	return mw.next.SendEmail(ctx, ver)
}

func LoggingMiddleware(log *slog.Logger) Middleware {
	return func(s Service) Service {
		return &loggingMiddleware{
			next:   s,
			logger: log,
		}
	}
}
