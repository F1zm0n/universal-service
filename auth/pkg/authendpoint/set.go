package authendpoint

import (
	"context"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/sony/gobreaker"

	"github.com/F1zm0n/uni-auth/pkg/authservice"
)

type Set struct {
	LoginEndpoint    endpoint.Endpoint
	RegisterEndpoint endpoint.Endpoint
}

func New(svc authservice.Service, logger log.Logger) Set {
	var loginEndpoint endpoint.Endpoint
	{
		loginEndpoint = makeLoginEndpoint(svc)
		loginEndpoint = circuitbreaker.Gobreaker(
			gobreaker.NewCircuitBreaker(gobreaker.Settings{}),
		)(
			loginEndpoint,
		)
		loginEndpoint = LoggingMiddleware(log.With(logger, "method", "login"))(loginEndpoint)
	}

	var registerEndpoint endpoint.Endpoint
	{
		registerEndpoint = makeRegisterEndpoint(svc)
		registerEndpoint = circuitbreaker.Gobreaker(
			gobreaker.NewCircuitBreaker(gobreaker.Settings{}),
		)(
			registerEndpoint,
		)
		registerEndpoint = LoggingMiddleware(
			log.With(logger, "method", "register"),
		)(
			registerEndpoint,
		)
	}
	return Set{
		RegisterEndpoint: registerEndpoint,
		LoginEndpoint:    loginEndpoint,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Err   error  `json:"error"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Err error `json:"error"`
}

func (s Set) Login(ctx context.Context, user authservice.User) (string, error) {
	resp, err := s.LoginEndpoint(
		ctx,
		LoginRequest{Email: user.Email, Password: user.Password},
	)
	if err != nil {
		return "", err
	}
	response := resp.(LoginResponse)
	return response.Token, response.Err
}

var (
	_ endpoint.Failer = LoginResponse{}
	_ endpoint.Failer = RegisterResponse{}
)

func (s Set) Register(ctx context.Context, user authservice.User) error {
	resp, err := s.RegisterEndpoint(
		ctx,
		RegisterRequest{Email: user.Email, Password: user.Password},
	)
	if err != nil {
		return err
	}
	response := resp.(RegisterResponse)
	return response.Err
}

func makeLoginEndpoint(s authservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(LoginRequest)
		user := authservice.User{
			Email:    req.Email,
			Password: req.Password,
		}
		tok, err := s.Login(ctx, user)
		return LoginResponse{Token: tok, Err: err}, nil
	}
}

func makeRegisterEndpoint(s authservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RegisterRequest)
		user := authservice.User{
			Email:    req.Email,
			Password: req.Password,
		}
		err = s.Register(ctx, user)
		return RegisterResponse{Err: err}, nil
	}
}

func (r LoginResponse) Failed() error    { return r.Err }
func (r RegisterResponse) Failed() error { return r.Err }
