package mailendpoint

import (
	"context"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/sony/gobreaker"

	"github.com/F1zm0n/universal-mailer/pkg/mailservice"
)

type Set struct {
	EmailEndpoint  endpoint.Endpoint
	VerifyEndpoint endpoint.Endpoint
}

func New(svc mailservice.Service, logger log.Logger) Set {
	var emailEndpoint endpoint.Endpoint
	{
		emailEndpoint = MakeEmailEndpoint(svc)
		emailEndpoint = circuitbreaker.Gobreaker(
			gobreaker.NewCircuitBreaker(gobreaker.Settings{}),
		)(
			emailEndpoint,
		)
		emailEndpoint = LoggingMiddleware(logger)(emailEndpoint)
	}
	var verifyEndpoint endpoint.Endpoint
	{
		verifyEndpoint = MakeVerifyEndpoint(svc)
		verifyEndpoint = circuitbreaker.Gobreaker(
			gobreaker.NewCircuitBreaker(gobreaker.Settings{}),
		)(
			verifyEndpoint,
		)
		verifyEndpoint = LoggingMiddleware(logger)(verifyEndpoint)
	}
	return Set{
		EmailEndpoint:  emailEndpoint,
		VerifyEndpoint: verifyEndpoint,
	}
}

func (s Set) SendEmail(ctx context.Context, ver mailservice.VerDto) error {
	resp, err := s.EmailEndpoint(ctx, EmailRequest{Email: ver.Email, Password: ver.Password})
	if err != nil {
		return err
	}
	response := resp.(EmailResponse)
	return response.Err
}

func (s Set) VerifyMail(ctx context.Context, id uuid.UUID) error {
	resp, err := s.VerifyEndpoint(ctx, VerifyRequest{VerId: id})
	if err != nil {
		return err
	}
	response := resp.(VerifyResponse)
	return response.Err
}

func MakeEmailEndpoint(s mailservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(EmailRequest)
		dto := mailservice.VerDto{
			Email:    req.Email,
			Password: req.Password,
		}
		err = s.SendEmail(ctx, dto)
		return EmailResponse{Err: err}, nil
	}
}

func MakeVerifyEndpoint(s mailservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(VerifyRequest)
		err = s.VerifyMail(ctx, req.VerId)
		return VerifyResponse{Err: err}, nil
	}
}

var (
	_ endpoint.Failer = EmailResponse{}
	_ endpoint.Failer = VerifyResponse{}
)

type EmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type EmailResponse struct {
	Err error `json:"error"`
}

// Failed implements endpoint.Failer.
func (e EmailResponse) Failed() error {
	return e.Err
}

type (
	VerifyRequest struct {
		VerId uuid.UUID `json:"ver_id"`
	}
	VerifyResponse struct {
		Err error `json:"error"`
	}
)

// Failed implements endpoint.Failer.
func (v VerifyResponse) Failed() error {
	return v.Err
}
