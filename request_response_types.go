package jsonrpc

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// Version defines the version of the JSON RPC implementation
	Version string = "2.0"

	// ContentType defines the content type to be served.
	ContentType string = "application/json; charset=utf-8"
)

// Request defines a JSON RPC request from the spec
// http://www.jsonrpc.org/specification#request_object
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      *RequestID      `json:"id"`
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

// UnmarshalJSON implements json.Unmarshaler
func (id *RequestID) UnmarshalJSON(b []byte) error {
	id.intError = json.Unmarshal(b, &id.intValue)
	id.floatError = json.Unmarshal(b, &id.floatValue)
	id.stringError = json.Unmarshal(b, &id.stringValue)
	if id.intError != nil && id.floatError != nil && id.stringError != nil {
		return fmt.Errorf("Unknown value for RequestID(%s)", b)
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
