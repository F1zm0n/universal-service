package prodtransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	"github.com/F1zm0n/universal-producer/pkg/prodendpoint"
	"github.com/F1zm0n/universal-producer/pkg/prodservice"
)

func NewHTTPHandler(endpoints prodendpoint.Set, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}
	m := http.NewServeMux()
	m.Handle("/mail", httptransport.NewServer(
		endpoints.MailEndpoint,
		decodeHTTPMailRequest,
		encodeHTTPGenericResponse,
		options...,
	))
	return m
}

func NewHTTPClient(instance string, logger log.Logger) (prodservice.Service, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	var options []httptransport.ClientOption

	var mailEndpoint endpoint.Endpoint
	{
		mailEndpoint = httptransport.NewClient(
			http.MethodPost,
			copyURL(u, `/mail`),
			encodeHTTPGenericRequest,
			decodeHTTPMailResponse,
			options...,
		).Endpoint()
		mailEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Mail",
			Timeout: 30 * time.Second,
		}))(mailEndpoint)
	}

	return prodendpoint.Set{
		MailEndpoint: mailEndpoint,
	}, nil
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

func err2code(err error) int {
	switch err {
	case prodservice.ErrProducingToKafka:
		return http.StatusInternalServerError
	}
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

func decodeHTTPMailRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req prodendpoint.MailRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPMailResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp prodendpoint.MailResponse
	err := json.NewDecoder(r.Body).Decode(&resp)

	return resp, err
}

func encodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		fmt.Println("error")
		return err
	}

	r.Body = io.NopCloser(&buf)
	return nil
}

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
