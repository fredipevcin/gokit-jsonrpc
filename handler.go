package jsonrpc

import (
	"context"
	"encoding/json"
	"net/http"
)

// Handlerer is the interface that provides method for serving JSON-RPC
type Handlerer interface {
	ServeJSONRPC(ctx context.Context, requestHeader http.Header, params json.RawMessage) (response interface{}, responseHeader http.Header, err error)
}
