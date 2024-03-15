package authtransport

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/log"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"

	authv1 "github.com/F1zm0n/uni-auth/pb"
	"github.com/F1zm0n/uni-auth/pkg/authendpoint"
	"github.com/F1zm0n/uni-auth/pkg/authservice"
)

type grpcServer struct {
	login    grpctransport.Handler
	register grpctransport.Handler
	authv1.UnimplementedAuthServiceServer
}

func NewGRPCServer(endpoints authendpoint.Set, logger log.Logger) authv1.AuthServiceServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	return &grpcServer{
		login: grpctransport.NewServer(
			endpoints.LoginEndpoint,
			decodeGRPCLoginRequest,
			encodeGRPCLoginResponse,
			options...,
		),
		register: grpctransport.NewServer(
			endpoints.RegisterEndpoint,
			decodeGRPCRegisterRequest,
			encodeGRPCRegisterResponse,
			options...,
		),
	}
}

func (s *grpcServer) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	_, rep, err := s.login.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*authv1.LoginResponse), nil
}

func (s *grpcServer) Register(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	_, rep, err := s.register.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*authv1.RegisterResponse), nil
}

func NewGRPCClient(conn *grpc.ClientConn, logger log.Logger) authservice.Service {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))

	var options []grpctransport.ClientOption

	var loginEndpoint endpoint.Endpoint
	{
		loginEndpoint = grpctransport.NewClient(
			conn,
			"pb.Auth",
			"Login",
			encodeGRPCLoginRequest,
			decodeGRPCLoginResponse,
			authv1.LoginResponse{},
			options...,
		).Endpoint()
		loginEndpoint = limiter(loginEndpoint)
		loginEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Login",
			Timeout: 30 * time.Second,
		}))(loginEndpoint)
	}

	var registerEndpoint endpoint.Endpoint
	{
		registerEndpoint = grpctransport.NewClient(
			conn,
			"pb.Auth",
			"Register",
			encodeGRPCRegisterRequest,
			decodeGRPCRegisterResponse,
			authv1.RegisterResponse{},
			options...,
		).Endpoint()
		registerEndpoint = limiter(registerEndpoint)
		registerEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Register",
			Timeout: 30 * time.Second,
		}))(registerEndpoint)
	}
	return authendpoint.Set{
		LoginEndpoint:    loginEndpoint,
		RegisterEndpoint: registerEndpoint,
	}
}

func decodeGRPCRegisterRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*authv1.RegisterRequest)
	return authendpoint.RegisterRequest{Email: req.GetEmail(), Password: req.Password}, nil
}

func decodeGRPCRegisterResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*authv1.RegisterResponse)
	return authendpoint.RegisterResponse{Err: stringToErr(reply.Err)}, nil
}

func encodeGRPCRegisterRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(authendpoint.RegisterRequest)
	return &authv1.RegisterRequest{Email: req.Email, Password: req.Password}, nil
}

func encodeGRPCRegisterResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(authendpoint.RegisterResponse)
	return &authv1.RegisterResponse{Err: errorToString(resp.Err)}, nil
}

func decodeGRPCLoginResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*authv1.LoginResponse)
	return authendpoint.LoginResponse{Err: stringToErr(reply.Err), Token: reply.Token}, nil
}

func decodeGRPCLoginRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*authv1.LoginRequest)
	return authendpoint.LoginRequest{Email: req.Email, Password: req.Password}, nil
}

func encodeGRPCLoginRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(authendpoint.LoginRequest)
	return &authv1.LoginRequest{Email: req.Email, Password: req.Password}, nil
}

func encodeGRPCLoginResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(authendpoint.LoginResponse)
	return &authv1.LoginResponse{Err: errorToString(resp.Err), Token: resp.Token}, nil
}

func stringToErr(s string) error {
	if s == "" {
		return nil
	}
	return errors.New(s)
}

func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
