package authtransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/sony/gobreaker"

	"github.com/F1zm0n/uni-auth/pkg/authendpoint"
	"github.com/F1zm0n/uni-auth/pkg/authservice"
)

func NewHTTPServer(endpoints authendpoint.Set, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}
	m := http.NewServeMux()
	m.Handle("/login", httptransport.NewServer(
		endpoints.LoginEndpoint,
		decodeHTTPLoginRequest,
		encodeHTTPGenericResponse,
		options...,
	))
	m.Handle("/register", httptransport.NewServer(
		endpoints.RegisterEndpoint,
		decodeHTTPRegisterRequest,
		encodeHTTPGenericResponse,
		options...,
	))
	return m
}

func NewHTTPClient(instance string, log log.Logger) (authservice.Service, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	var options []httptransport.ClientOption

	var loginEndpoint endpoint.Endpoint
	{
		loginEndpoint = httptransport.NewClient(
			http.MethodGet,
			copyURL(u, "/login"),
			encodeHTTPGenericRequest,
			decodeHTTPLoginResponse,
			options...,
		).Endpoint()
		loginEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Login",
			Timeout: 30 * time.Second,
		}))(loginEndpoint)
	}

	var registerEndpoint endpoint.Endpoint
	{
		registerEndpoint = httptransport.NewClient(
			http.MethodPost,
			copyURL(u, "/register"),
			encodeHTTPGenericRequest,
			decodeHTTPRegisterResponse,
			options...,
		).Endpoint()
		registerEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Register",
			Timeout: 30 * time.Second,
		}))(registerEndpoint)
	}

	return authendpoint.Set{
		RegisterEndpoint: registerEndpoint,
		LoginEndpoint:    loginEndpoint,
	}, nil
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

func decodeHTTPLoginRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req authendpoint.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPRegisterRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req authendpoint.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPLoginResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp authendpoint.LoginResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func decodeHTTPRegisterResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp authendpoint.RegisterResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func encodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = io.NopCloser(&buf)
	return nil
}

// encodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func encodeHTTPGenericResponse(
	ctx context.Context,
	w http.ResponseWriter,
	response interface{},
) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

func err2code(err error) int {
	switch err {
	case authservice.ErrInvalidCreds:
		return http.StatusUnauthorized
	case authservice.ErrGeneratingToken, authservice.ErrInsertingUser:
		return http.StatusInternalServerError
	case authservice.ErrWrongEmailFmt, authservice.ErrWrongPassFmt:
		return http.StatusBadRequest
	case authservice.ErrUserAlreadyExists:
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}

func err2info(err error) error {
	switch err {
	case authservice.ErrGeneratingToken, authservice.ErrInsertingUser:
		return errors.New("internal server error")
	case authservice.ErrInvalidCreds:
		return authservice.ErrInvalidCreds
	case authservice.ErrWrongPassFmt:
		return authservice.ErrWrongPassFmt
	case authservice.ErrWrongEmailFmt:
		return authservice.ErrWrongEmailFmt
	case authservice.ErrUserAlreadyExists:
		return authservice.ErrUserAlreadyExists
	}
	return errors.New("internal server error")
}

func errorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}
