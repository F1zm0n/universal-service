package mailtransport

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

	"github.com/F1zm0n/universal-mailer/pkg/mailendpoint"
	"github.com/F1zm0n/universal-mailer/pkg/mailservice"
)

func NewHTTPHandler(endpoints mailendpoint.Set, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}
	m := http.NewServeMux()
	m.Handle("/mail", httptransport.NewServer(
		endpoints.EmailEndpoint,
		decodeHTTPEmailRequest,
		encodeHTTPGenericResponse,
		options...,
	))
	m.Handle("/verify", httptransport.NewServer(
		endpoints.VerifyEndpoint,
		decodeHTTPVerifyRequest,
		encodeHTTPGenericResponse,
		options...,
	))
	return m
}

func NewHTTPClient(instance string, logger log.Logger) (mailservice.Service, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	var options []httptransport.ClientOption
	var emailEndpoint endpoint.Endpoint
	{
		emailEndpoint = httptransport.NewClient(
			http.MethodPost,
			copyURL(u, "/mail"),
			encodeHTTPGenericRequest,
			decodeHTTPEmailResponse,
			options...,
		).Endpoint()

		emailEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Mail",
			Timeout: 30 * time.Second,
		}))(emailEndpoint)
	}
	var verifyEndpoint endpoint.Endpoint
	{
		verifyEndpoint = httptransport.NewClient(
			http.MethodGet,
			copyURL(u, "/verify"),
			encodeHTTPGenericRequest,
			decodeHTTPVerifyResponse,
			options...,
		).Endpoint()
		verifyEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Verify",
			Timeout: 30 * time.Second,
		}))(verifyEndpoint)

	}
	return mailendpoint.Set{
		EmailEndpoint:  emailEndpoint,
		VerifyEndpoint: verifyEndpoint,
	}, nil
}

func decodeHTTPEmailRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req mailendpoint.EmailRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPEmailResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp mailendpoint.EmailResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func decodeHTTPVerifyResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp mailendpoint.VerifyResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func decodeHTTPVerifyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// id := r.URL.Query().Get("id")
	// verId, err := uuid.Parse(id)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// req := mailendpoint.VerifyRequest{
	// 	VerId: verId,
	// }
	var req mailendpoint.VerifyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

func err2code(_ error) int {
	return http.StatusInternalServerError
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

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

// encodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
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
