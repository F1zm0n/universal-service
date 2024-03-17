package prodendpoint

import (
	"context"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/sony/gobreaker"

	"github.com/F1zm0n/universal-producer/pkg/prodservice"
)

type Set struct {
	MailEndpoint     endpoint.Endpoint
	RegisterEndpoint endpoint.Endpoint
	VerEndpoint      endpoint.Endpoint
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

	var registerEndpoint endpoint.Endpoint
	{
		registerEndpoint = MakeRegisterEndpoint(svc)
		registerEndpoint = circuitbreaker.Gobreaker(
			gobreaker.NewCircuitBreaker(gobreaker.Settings{}),
		)(
			registerEndpoint,
		)
		registerEndpoint = LoggingMiddleware(logger)(registerEndpoint)
	}

	var verEndpoint endpoint.Endpoint
	{
		verEndpoint = MakeVerEndpoint(svc)
		verEndpoint = circuitbreaker.Gobreaker(
			gobreaker.NewCircuitBreaker(gobreaker.Settings{}),
		)(
			verEndpoint,
		)
		verEndpoint = LoggingMiddleware(logger)(verEndpoint)
	}
	return Set{
		MailEndpoint:     mailEndpoint,
		RegisterEndpoint: registerEndpoint,
		VerEndpoint:      verEndpoint,
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

func (s Set) ProduceVer(ctx context.Context, verId uuid.UUID) error {
	resp, err := s.VerEndpoint(ctx, VerRequest{VerId: verId})
	if err != nil {
		return err
	}
	response := resp.(VerResponse)
	return response.Err
}

func (s Set) ProduceRegister(ctx context.Context, password, email string) error {
	resp, err := s.MailEndpoint(ctx, RegisterRequest{Email: email, Password: password})
	if err != nil {
		return err
	}
	response := resp.(RegisterResponse)
	return response.Err
}

func MakeMailEndpoint(s prodservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(MailRequest)
		err = s.ProduceMail(ctx, req.Email, req.Password)
		return MailResponse{Err: err}, nil
	}
}

func MakeVerEndpoint(s prodservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(VerRequest)
		err = s.ProduceVer(ctx, req.VerId)
		return VerResponse{Err: err}, nil
	}
}

func MakeRegisterEndpoint(s prodservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RegisterRequest)
		err = s.ProduceRegister(ctx, req.Email, req.Password)
		return RegisterResponse{Err: err}, nil
	}
}

type MailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MailResponse struct {
	Err error `json:"error"`
}
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RegisterResponse struct {
	Err error `json:"error"`
}

type VerRequest struct {
	VerId uuid.UUID `json:"ver_id"`
}
type VerResponse struct {
	Err error `json:"error"`
}

var (
	_ endpoint.Failer = MailResponse{}
	_ endpoint.Failer = RegisterResponse{}
	_ endpoint.Failer = VerResponse{}
)

func (r RegisterResponse) Failed() error {
	return r.Err
}

func (v VerResponse) Failed() error {
	return v.Err
}

func (m MailResponse) Failed() error {
	return m.Err
}
