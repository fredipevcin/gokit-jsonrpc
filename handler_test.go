package jsonrpc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	jsonrpc "github.com/fredipevcin/gokit-jsonrpc"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

func TestHandlerBadDecode(t *testing.T) {
	handler := jsonrpc.NewHandler(
		endpoint.Nop,
		func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, errors.New("dang") },
		func(context.Context, interface{}) (json.RawMessage, error) { return nil, nil },
	)

	ctx := context.Background()
	requestHeader := http.Header{}
	params := json.RawMessage{}

	_, _, err := handler.ServeJSONRPC(ctx, requestHeader, params)

	if err == nil {
		t.Error("Expect err to be nil")
	}

	if got, expect := err.Error(), "dang"; got != expect {
		t.Errorf("got %s, expect %s", got, expect)
	}
}

func TestHandlerBadEndpoint(t *testing.T) {
	handler := jsonrpc.NewHandler(
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, errors.New("dang") },
		func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, interface{}) (json.RawMessage, error) { return nil, nil },
	)

	ctx := context.Background()
	requestHeader := http.Header{}
	params := json.RawMessage{}

	_, _, err := handler.ServeJSONRPC(ctx, requestHeader, params)

	if err == nil {
		t.Error("Expect err to be nil")
	}

	if got, expect := err.Error(), "dang"; got != expect {
		t.Errorf("got %s, expect %s", got, expect)
	}
}

func TestHandlerBadEncode(t *testing.T) {
	handler := jsonrpc.NewHandler(
		endpoint.Nop,
		func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, interface{}) (json.RawMessage, error) { return nil, errors.New("dang") },
	)

	ctx := context.Background()
	requestHeader := http.Header{}
	params := json.RawMessage{}

	_, _, err := handler.ServeJSONRPC(ctx, requestHeader, params)

	if err == nil {
		t.Error("Expect err to be nil")
	}

	if got, expect := err.Error(), "dang"; got != expect {
		t.Errorf("got %s, expect %s", got, expect)
	}
}

func TestMultipleHandlerBefore(t *testing.T) {
	var done = make(chan struct{})

	handler := jsonrpc.NewHandler(
		endpoint.Nop,
		func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, interface{}) (json.RawMessage, error) { return nil, nil },
		jsonrpc.HandlerBefore(func(ctx context.Context, headers http.Header) context.Context {
			ctx = context.WithValue(ctx, "one", 1)

			return ctx
		}),
		jsonrpc.HandlerBefore(func(ctx context.Context, headers http.Header) context.Context {
			if _, ok := ctx.Value("one").(int); !ok {
				t.Error("Value was not set properly when multiple HandlerBefore are used")
			}

			close(done)
			return ctx
		}),
	)

	handler.ServeJSONRPC(context.Background(), http.Header{}, json.RawMessage{})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for finalizer")
	}
}

func TestMultipleHandlerAfter(t *testing.T) {

	var done = make(chan struct{})

	handler := jsonrpc.NewHandler(
		endpoint.Nop,
		func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, interface{}) (json.RawMessage, error) { return nil, nil },
		jsonrpc.HandlerAfter(func(ctx context.Context, headers http.Header) context.Context {
			ctx = context.WithValue(ctx, "one", 1)

			return ctx
		}),
		jsonrpc.HandlerAfter(func(ctx context.Context, headers http.Header) context.Context {
			if _, ok := ctx.Value("one").(int); !ok {
				t.Error("Value was not set properly when multiple HandlerAfter are used")
			}

			close(done)
			return ctx
		}),
	)

	handler.ServeJSONRPC(context.Background(), http.Header{}, json.RawMessage{})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for finalizer")
	}
}

func TestHandlerLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := log.NewLogfmtLogger(buf)

	handler := jsonrpc.NewHandler(
		endpoint.Nop,
		func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, interface{}) (json.RawMessage, error) { return nil, errors.New("dang") },
		jsonrpc.HandlerErrorLogger(logger),
	)

	handler.ServeJSONRPC(context.Background(), http.Header{}, json.RawMessage{})

	if got, expect := strings.TrimSpace(buf.String()), "err=dang"; got != expect {
		t.Errorf("got %s, expect %s", got, expect)
	}
}
