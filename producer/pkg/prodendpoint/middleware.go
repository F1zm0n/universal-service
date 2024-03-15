package prodendpoint

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
)

func LoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(start time.Time) {
				logger.Log(
					"transport_error", err,
					"took", time.Since(start),
				)
			}(time.Now())
			return e(ctx, request)
		}
	}
}
