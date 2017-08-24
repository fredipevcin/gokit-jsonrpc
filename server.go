package jsonrpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
)

// Handlers maps the method to the proper handler
type Handlers map[string]Handlerer

// ServerOption sets an optional parameter for servers
type ServerOption func(*Server)

// ServerErrorEncoder is used to encode errors to the http.ResponseWriter
// whenever they're encountered in the processing of a request. Clients can
// use this to provide custom error formatting and response codes. By default,
// errors will be written with the DefaultErrorEncoder.
func ServerErrorEncoder(ee httptransport.ErrorEncoder) ServerOption {
	return func(s *Server) { s.errorEncoder = ee }
}

// NewServer constructs a new Server, which implements http.Handler
func NewServer(
	sh Handlers,
	options ...ServerOption,
) *Server {
	s := &Server{
		sh:           sh,
		errorEncoder: DefaultErrorEncoder,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// Server wraps an list of handlers and implements http.Handler
type Server struct {
	sh           Handlers
	errorEncoder httptransport.ErrorEncoder
}

// ServeHTTP implements http.Handler
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must POST\n")
		return
	}

	ctx := r.Context()
	ctx = httptransport.PopulateRequestContext(ctx, r)

	// Decode the body into an  object
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.errorEncoder(ctx, NewError(ParseError), w)
		return
	}

	// has to be specific version and method should not be empty
	if req.JSONRPC != Version || req.Method == "" {
		s.errorEncoder(ctx, NewError(InvalidRequestError), w)
		return
	}

	// Get the endpoint and codecs from the map using the method
	// defined in the JSON  object
	srv, ok := s.sh[req.Method]
	if !ok {
		s.errorEncoder(ctx, NewError(MethodNotFoundError), w)
		return
	}

	resp, respHeaders, err := srv.ServeJSONRPC(ctx, r.Header, req.Params)
	if err != nil {
		s.errorEncoder(ctx, err, w)
		return
	}

	res := Response{
		RespHeaders: respHeaders,
		JSONRPC:     Version,
		ID:          req.ID,
	}

	res.Result = resp

	httptransport.EncodeJSONResponse(ctx, w, res)
	// err = httptransport.EncodeJSONResponse(ctx, w, res)
	// if err != nil {
	// 	s.errorEncoder(ctx, err, w)
	// 	return
	// }
}

// DefaultErrorEncoder writes the error to the ResponseWriter,
// as a json-rpc error response, with an InternalError status code.
// The Error() string of the error will be used as the response error message.
// If the error implements ErrorCoder, the provided code will be set on the
// response error.
// If the error implements Headerer, the given headers will be set.
func DefaultErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", ContentType)
	e := NewError(InternalError, err.Error())
	if sc, ok := err.(ErrorCoder); ok {
		e.Code = sc.ErrorCode()
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		JSONRPC: Version,
		Error:   &e,
	})
}

// StatusCoder is checked by DefaultErrorEncoder. If an error value implements
// StatusCoder, the StatusCode will be used when encoding the error. By default,
// StatusInternalServerError (500) is used.
type StatusCoder interface {
	StatusCode() int
}

// Headerer is checked by DefaultErrorEncoder. If an error value implements
// Headerer, the provided headers will be applied to the response writer, after
// the Content-Type is set.
type Headerer interface {
	Headers() http.Header
}