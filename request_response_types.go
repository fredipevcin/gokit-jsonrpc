package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const (
	// Version defines the version of the JSON RPC implementation
	Version string = "2.0"

	// ContentType defines the content type to be served.
	ContentType string = "application/json; charset=utf-8"
)

// ErrParsingRequestID is used when request id cannot be parsed
var ErrParsingRequestID = errors.New("Unknown value for RequestID")

// Request defines a JSON RPC request from the spec
// http://www.jsonrpc.org/specification#request_object
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      *RequestID      `json:"id"`
}

// Validate request
func (r *Request) Validate() error {
	// A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0"
	if r.JSONRPC != Version {
		return NewError(InvalidRequestError)
	}

	// An identifier established by the Client that MUST contain a String, Number, or NULL value if included.
	// If it is not included it is assumed to be a notification.
	if r.ID != nil && r.ID.Error() != "" {
		return NewError(InvalidRequestError)
	}

	// A String containing the name of the method to be invoked.
	// Method names that begin with the word rpc followed by a period character (U+002E or ASCII 46) are reserved for
	// rpc-internal methods and extensions and MUST NOT be used for anything else.
	if r.Method == "" || strings.HasPrefix(r.Method, "rpc.") {
		return NewError(InvalidRequestError)
	}

	return nil
}

// RequestID defines a request ID that can be string, number, or null.
// An identifier established by the Client that MUST contain a String,
// Number, or NULL value if included.
// If it is not included it is assumed to be a notification.
// The value SHOULD normally not be Null and
// Numbers SHOULD NOT contain fractional parts.
type RequestID struct {
	intValue    int
	intError    error
	floatValue  float32
	floatError  error
	stringValue string
	stringError error
}

// Error implements errors interface
func (id *RequestID) Error() string {
	if id.intError == nil || id.floatError == nil || id.stringError == nil {
		return ""
	}

	return ErrParsingRequestID.Error()
}

// UnmarshalJSON implements json.Unmarshaler
func (id *RequestID) UnmarshalJSON(b []byte) error {
	id.intError = json.Unmarshal(b, &id.intValue)
	id.floatError = json.Unmarshal(b, &id.floatValue)
	id.stringError = json.Unmarshal(b, &id.stringValue)
	if id.intError != nil && id.floatError != nil && id.stringError != nil {
		return ErrParsingRequestID
	}
	return nil
}

// MarshalJSON implements json.Marshaler
func (id *RequestID) MarshalJSON() ([]byte, error) {
	if id.intError == nil && id.floatError == nil && id.stringError == nil {
		return []byte("null"), nil
	}

	if id.intError == nil {
		return json.Marshal(id.intValue)
	}
	if id.floatError == nil {
		return json.Marshal(id.floatValue)
	}

	return json.Marshal(id.stringValue)
}

// Int returns the ID as an integer value.
// An error is returned if the ID can't be treated as an int.
func (id *RequestID) Int() (int, error) {
	return id.intValue, id.intError
}

// Float32 returns the ID as a float value.
// An error is returned if the ID can't be treated as an float.
func (id *RequestID) Float32() (float32, error) {
	return id.floatValue, id.floatError
}

// String returns the ID as a string value.
// An error is returned if the ID can't be treated as an string.
func (id *RequestID) String() (string, error) {
	return id.stringValue, id.stringError
}

// Response defines a JSON RPC response from the spec
// http://www.jsonrpc.org/specification#response_object
type Response struct {
	JSONRPC     string      `json:"jsonrpc"`
	Result      interface{} `json:"result,omitempty"`
	Error       *Error      `json:"error,omitempty"`
	ID          *RequestID  `json:"id,omitempty"`
	RespHeaders http.Header `json:"-"`
}

// Headers returns response headers
func (r Response) Headers() http.Header {
	return r.RespHeaders
}

// PopulateRequestContext is a RequestFunc that populates several values into
// the context from the JSONRPC request. Those values may be extracted using the
// corresponding ContextKey type in this package.
func PopulateRequestContext(ctx context.Context, r *Request) context.Context {
	for k, v := range map[contextKey]interface{}{
		ContextKeyRequestJSONRPC: r.JSONRPC,
		ContextKeyRequestMethod:  r.Method,
		ContextKeyRequestID:      r.ID,
	} {
		ctx = context.WithValue(ctx, k, v)
	}
	return ctx
}

type contextKey int

const (
	// ContextKeyRequestJSONRPC is populated in the context by PopulateRequestContext
	ContextKeyRequestJSONRPC contextKey = iota

	// ContextKeyRequestMethod is populated in the context by PopulateRequestContext
	ContextKeyRequestMethod

	// ContextKeyRequestID is populated in the context by PopulateRequestContext
	ContextKeyRequestID
)
