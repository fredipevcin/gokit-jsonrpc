package jsonrpc

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// HandlerRequestFunc may take information from an header and put it into a
// request context.
type HandlerRequestFunc func(context.Context, http.Header) context.Context

// HandlerResponseFunc may take information from a request context and use it to
// add to response header
type HandlerResponseFunc func(context.Context, http.Header) context.Context

// DecodeRequestFunc extracts a user-domain request object from params.
type DecodeRequestFunc func(context.Context, json.RawMessage) (request interface{}, err error)

// EncodeResponseFunc encodes the passed response object
type EncodeResponseFunc func(context.Context, interface{}) (response json.RawMessage, err error)

// Handler implements Handlerer interface
type Handler struct {
	e      endpoint.Endpoint
	dec    DecodeRequestFunc
	enc    EncodeResponseFunc
	before []HandlerRequestFunc
	after  []HandlerResponseFunc
	logger log.Logger
}

// NewHandler constructs a new handler, which implements jsonrcp.Handlerer and wraps
// the provided endpoint.
func NewHandler(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	options ...HandlerOption,
) *Handler {
	s := &Handler{
		e:      e,
		dec:    dec,
		enc:    enc,
		logger: log.NewNopLogger(),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// HandlerOption sets an optional parameter for servers.
type HandlerOption func(*Handler)

// HandlerBefore functions are executed on the request object
func HandlerBefore(before ...HandlerRequestFunc) HandlerOption {
	return func(s *Handler) { s.before = append(s.before, before...) }
}

// HandlerAfter functions are executed on the response after the
// endpoint is invoked, but before anything is written to the client.
func HandlerAfter(after ...HandlerResponseFunc) HandlerOption {
	return func(s *Handler) { s.after = append(s.after, after...) }
}

// HandlerErrorLogger is used to log non-terminal errors. By default, no errors
// are logged. This is intended as a diagnostic measure. Finer-grained control
// of error handling, including logging in more detail, should be performed in a
// custom ServerErrorEncoder or ServerFinalizer, both of which have access to
// the context.
func HandlerErrorLogger(logger log.Logger) HandlerOption {
	return func(s *Handler) { s.logger = logger }
}

// ServeJSONRPC implements Handlerer
func (s Handler) ServeJSONRPC(ctx context.Context, requestHeader http.Header, params json.RawMessage) (responseParams json.RawMessage, responseHeader http.Header, err error) {
	for _, f := range s.before {
		ctx = f(ctx, requestHeader)
	}

	request, err := s.dec(ctx, params)
	if err != nil {
		s.logger.Log("err", err)
		return nil, nil, err
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.logger.Log("err", err)
		return nil, nil, err
	}

	responseHeader = http.Header{}
	for _, f := range s.after {
		ctx = f(ctx, responseHeader)
	}

	// Encode the response from the Endpoint
	responseParams, err = s.enc(ctx, response)
	if err != nil {
		s.logger.Log("err", err)
		return nil, nil, err
	}

	return responseParams, responseHeader, nil
}
