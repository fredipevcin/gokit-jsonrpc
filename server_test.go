package jsonrpc_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jsonrpc "github.com/fredipevcin/gokit-jsonrpc"
)

const testMethodName = "test"

type HandlererFunc func(ctx context.Context, requestHeader http.Header, params json.RawMessage) (response interface{}, responseHeader http.Header, err error)

func (h HandlererFunc) ServeJSONRPC(ctx context.Context, requestHeader http.Header, params json.RawMessage) (response interface{}, responseHeader http.Header, err error) {
	return h(ctx, requestHeader, params)
}

func nopHandler(ctx context.Context, requestHeader http.Header, params json.RawMessage) (response interface{}, responseHeader http.Header, err error) {
	return nil, nil, nil
}

func testServer(r *http.Request, hander jsonrpc.Handlerer) (*httptest.ResponseRecorder, error) {
	var retErr error
	server := jsonrpc.NewServer(
		jsonrpc.Handlers{
			testMethodName: hander,
		},
		jsonrpc.ServerErrorEncoder(func(_ context.Context, err error, w http.ResponseWriter) {
			retErr = err
		}),
	)

	rw := httptest.NewRecorder()
	server.ServeHTTP(rw, r)

	return rw, retErr
}
func TestServerInvalidMethod(t *testing.T) {

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	h := HandlererFunc(nopHandler)
	rw, err := testServer(r, h)

	if got, expect := rw.Code, http.StatusMethodNotAllowed; got != expect {
		t.Errorf("Expected response code %d, got %d", expect, expect)
	}

	if err != nil {
		t.Errorf("Expecting error to be nil, got: %s", err)
	}

	body := strings.TrimSpace(rw.Body.String())
	if body != "405 must POST" {
		t.Errorf("Unexpected response body %s", body)
	}
}

func TestServerRequestBodyCannotBeParsed(t *testing.T) {
	cases := []string{
		"notjson",
		`{"id":true}`,
		`{"method":1234}`,
		`{"jsonrpc":1234}`,
	}

	for _, c := range cases {
		r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(c))
		h := HandlererFunc(nopHandler)
		_, err := testServer(r, h)

		if err == nil {
			t.Fatalf("TC(%s) Expecting error, got nil", c)
		}

		rpcErr, ok := err.(jsonrpc.Errorer)
		if !ok {
			t.Fatalf("TC(%s) Expected err implements jsonrpc.Erroer for type %T", c, err)
		}

		if got, expect := rpcErr.ErrorCode(), jsonrpc.ParseError; got != expect {
			t.Errorf(" TC(%s) Expecting error code %d, got %d", c, expect, got)
		}

	}
}

func TestServerRequestIsNotValid(t *testing.T) {
	cases := []string{
		`{"id":"1234"}`,
		`{"method":"a.b"}`,
		`{"method":"some_method"}`,
		`{"jsonrpc":"1","method":"some_method"}`,
		`{"jsonrpc":"2.1","method":"some_method"}`,
		fmt.Sprintf(`{"jsonrpc":"2","method":"%s"}`, testMethodName),
	}
	for _, c := range cases {
		r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(c))
		h := HandlererFunc(nopHandler)
		_, err := testServer(r, h)

		if err == nil {
			t.Fatalf("TC(%s) Expecting error, got nil", c)
		}

		rpcErr, ok := err.(jsonrpc.Errorer)
		if !ok {
			t.Fatalf("TC(%s) Expected err implements jsonrpc.Erroer for type %T", c, err)
		}

		if got, expect := rpcErr.ErrorCode(), jsonrpc.InvalidRequestError; got != expect {
			t.Errorf(" TC(%s) Expecting error code %d, got %d", c, expect, got)
		}
	}
}

func TestServerMethodNotFound(t *testing.T) {
	cases := []string{
		`{"jsonrpc":"2.0","method":"some_method3"}`,
		`{"jsonrpc":"2.0","method":"some_method1","id":1234}`,
		`{"jsonrpc":"2.0","method":"some_method2","params":{"a":"b"}}`,
	}
	for _, c := range cases {
		r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(c))
		h := HandlererFunc(nopHandler)
		_, err := testServer(r, h)

		if err == nil {
			t.Fatalf("TC(%s) Expecting error, got nil", c)
		}

		rpcErr, ok := err.(jsonrpc.Errorer)
		if !ok {
			t.Fatalf("TC(%s) Expected err implements jsonrpc.Erroer for type %T", c, err)
		}

		if got, expect := rpcErr.ErrorCode(), jsonrpc.MethodNotFoundError; got != expect {
			t.Errorf(" TC(%s) Expecting error code %d, got %d", c, expect, got)
		}
	}
}

func TestServerHandlerError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","params":{"a":"b"}}`, testMethodName)))
	h := HandlererFunc(func(ctx context.Context, requestHeader http.Header, params json.RawMessage) (response interface{}, responseHeader http.Header, err error) {
		return nil, nil, errors.New("oooh")
	})
	_, err := testServer(r, h)

	if err == nil {
		t.Fatal("Expecting error, got nil")
	}

	if err.Error() != "oooh" {
		t.Fatalf("Expecting error(oooh), got %s", err)
	}
}

func TestServerSuccess(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","params":{"a":"b"}}`, testMethodName)))
	r.Header.Set("X-Foo", "bar")

	h := HandlererFunc(func(ctx context.Context, requestHeader http.Header, params json.RawMessage) (response interface{}, responseHeader http.Header, err error) {

		if got, expected := string(params), `{"a":"b"}`; got != expected {
			t.Errorf("Expect params %s, got %s", expected, got)
		}

		if got, expected := requestHeader.Get("x-foo"), "bar"; got != expected {
			t.Errorf("Expect header.Get(x-foo) %s, got %s", expected, got)
		}

		hdr := http.Header{}
		hdr.Set("X-ReqId", "124")

		return "woohoo", hdr, nil
	})
	rw, err := testServer(r, h)

	if err != nil {
		t.Fatalf("Expecting error to be nil, got %s", err)
	}

	if got, expect := rw.Code, http.StatusOK; got != expect {
		t.Errorf("Expected response code %d, got %d", expect, expect)
	}

	if got, expect := rw.Header().Get("content-type"), "application/json; charset=utf-8"; got != expect {
		t.Errorf("Expected content-type '%s', got '%s'", got, expect)
	}
	if got, expect := rw.Header().Get("x-reqid"), "124"; got != expect {
		t.Errorf("Expected response header '%s', got '%s'", got, expect)
	}

	if got, expect := strings.TrimSpace(rw.Body.String()), `{"jsonrpc":"2.0","result":"woohoo"}`; got != expect {
		t.Errorf("Expected body '%s', got '%s'", got, expect)
	}
}

func TestDefaultErrorEncoderWithPredfinedErrors(t *testing.T) {
	rw := httptest.NewRecorder()
	err := jsonrpc.NewError(jsonrpc.InternalError)
	jsonrpc.DefaultErrorEncoder(context.Background(), err, rw)

	if got, expect := rw.Code, http.StatusOK; got != expect {
		t.Errorf("Expected response code %d, got %d", expect, expect)
	}
	if got, expect := strings.TrimSpace(rw.Body.String()), `{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal JSON-RPC error"}}`; got != expect {
		t.Errorf("Expected body '%s', got '%s'", got, expect)
	}
}
func TestDefaultErrorEncoderWithCustomMessages(t *testing.T) {
	rw := httptest.NewRecorder()
	err := jsonrpc.NewError(jsonrpc.InvalidParamsError, "Booya")
	jsonrpc.DefaultErrorEncoder(context.Background(), err, rw)

	if got, expect := rw.Code, http.StatusOK; got != expect {
		t.Errorf("Expected response code %d, got %d", expect, expect)
	}
	if got, expect := strings.TrimSpace(rw.Body.String()), `{"jsonrpc":"2.0","error":{"code":-32602,"message":"Booya"}}`; got != expect {
		t.Errorf("Expected body '%s', got '%s'", got, expect)
	}
}
