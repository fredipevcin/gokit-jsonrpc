package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
)

// Handlers maps the method to the proper handler
type Handlers map[string]Handlerer

// Set handler for given method, panics if method already exists
func (h Handlers) Set(method string, handler Handlerer) {
	if _, ok := h[method]; ok {
		panic(fmt.Sprintf("Handler for method %s already exists", method))
	}
	h[method] = handler
}

// Handlerer is the interface that provides method for serving JSON-RPC
type Handlerer interface {
	ServeJSONRPC(ctx context.Context, requestHeader http.Header, params json.RawMessage) (response json.RawMessage, responseHeader http.Header, err error)
}

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

	ctx = PopulateRequestContext(ctx, &req)

	if err := req.Validate(); err != nil {
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

	// notification
	if req.ID == nil {
		go srv.ServeJSONRPC(ctx, r.Header, req.Params)
		httptransport.EncodeJSONResponse(ctx, w, NotificationResponse{})
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

	var respErr Error
	err = json.Unmarshal(resp, &respErr)
	// try to find out if resp is an error
	if err == nil && respErr.ErrorCode() != 0 {
		// it has to set a pointer otherwise in Go 1.7 base64 encoded string is returned.
		// In Go 1.8 works as expected
		// Golang release notes 1.8: A RawMessage value now marshals the same as its pointer type.
		res.Error = &respErr
	} else {
		res.Result = &resp
	}

	httptransport.EncodeJSONResponse(ctx, w, res)
}

// DefaultErrorEncoder writes the error to the ResponseWriter,
// as a json-rpc error response, with an InternalError status code.
// The Error() string of the error will be used as the response error message.
// If the error implements ErrorCoder, the provided code will be set on the
// response error.
// If the error implements Headerer, the given headers will be set.
func DefaultErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", ContentType)
	e := NewError(InternalError)
	if te, ok := err.(ErrorCoder); ok {
		e.Code = te.ErrorCode()
	}

	if te, ok := err.(Errorer); ok {
		e.Message = te.Error()
	}

	w.WriteHeader(http.StatusOK)

	reqID, _ := ctx.Value(ContextKeyRequestID).(*RequestID)
	json.NewEncoder(w).Encode(Response{
		ID:      reqID,
		JSONRPC: Version,
		Error:   &e,
	})
}

// Headerer is checked by DefaultErrorEncoder. If an error value implements
// Headerer, the provided headers will be applied to the response writer, after
// the Content-Type is set.
type Headerer interface {
	Headers() http.Header
}
