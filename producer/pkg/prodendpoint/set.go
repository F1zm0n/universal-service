package prodendpoint

import (
	"context"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/sony/gobreaker"

	"github.com/F1zm0n/universal-producer/pkg/prodservice"
)

type Set struct {
	MailEndpoint endpoint.Endpoint
}

func New(svc prodservice.Service, logger log.Logger) Set {
	var mailEndpoint endpoint.Endpoint
	{
		mailEndpoint = MakeMailEndpoint(svc)
		mailEndpoint = circuitbreaker.Gobreaker(
			gobreaker.NewCircuitBreaker(gobreaker.Settings{}),
		)(
			mailEndpoint,
		)
		mailEndpoint = LoggingMiddleware(logger)(mailEndpoint)
	}
	return Set{
		MailEndpoint: mailEndpoint,
	}
}

func (s Set) ProduceMail(ctx context.Context, password, email string) error {
	resp, err := s.MailEndpoint(ctx, MailRequest{Email: email, Password: password})
	if err != nil {
		return err
	}
	response := resp.(MailResponse)
	return response.Err
}

func MakeMailEndpoint(s prodservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(MailRequest)
		err = s.ProduceMail(ctx, req.Email, req.Password)
		return MailResponse{Err: err}, nil
	}
}

type MailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MailResponse struct {
	Err error `json:"error"`
}

// Failed implements endpoint.Failer.
func (m MailResponse) Failed() error {
	return m.Err
}

var _ endpoint.Failer = MailResponse{}
