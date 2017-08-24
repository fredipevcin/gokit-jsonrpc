package jsonrpc

// Errorer describes methods for managing errors
type Errorer interface {
	Error() string
	ErrorCode() int
}

// Error defines a JSON RPC error that can be returned
// in a Response from the spec
// http://www.jsonrpc.org/specification#error_object
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements Errorer and error
func (e Error) Error() string {
	return e.Message
}

// ErrorCode implements Errorer
func (e Error) ErrorCode() int {
	return e.Code
}

const (
	// ParseError defines invalid JSON was received by the server.
	// An error occurred on the server while parsing the JSON text.
	ParseError int = -32700

	// InvalidRequestError defines the JSON sent is not a valid Request object.
	InvalidRequestError int = -32600

	// MethodNotFoundError defines the method does not exist / is not available.
	MethodNotFoundError int = -32601

	// InvalidParamsError defines invalid method parameter(s).
	InvalidParamsError int = -32602

	// InternalError defines a server error
	InternalError int = -32603
)

var errorMessage = map[int]string{
	ParseError:          "An error occurred on the server while parsing the JSON text",
	InvalidRequestError: "The JSON sent is not a valid Request object",
	MethodNotFoundError: "The method does not exist / is not available",
	InvalidParamsError:  "Invalid method parameter(s)",
	InternalError:       "Internal JSON-RPC error",
}

// NewError returns Error struct
func NewError(code int, message ...string) Error {
	msg := ErrorMessage(code)
	if len(message) > 0 {
		msg = message[0]
	}
	return Error{
		Code:    code,
		Message: msg,
	}
}

// ErrorMessage returns a message for the JSON RPC error code. It returns the empty
// string if the code is unknown.
func ErrorMessage(code int) string {
	return errorMessage[code]
}
