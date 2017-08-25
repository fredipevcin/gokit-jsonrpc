package jsonrpc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jsonrpc "github.com/fredipevcin/gokit-jsonrpc"
)

func addBody(method string, params interface{}) io.Reader {
	jsonParams, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	return strings.NewReader(fmt.Sprintf(`{"jsonrpc": "2.0", "method": "%s", "params": %s, "id": 1}`, method, jsonParams))
}

func TestHandlerInvalidParameters(t *testing.T) {
	jsonserver := jsonrpc.NewServer(jsonrpc.Handlers{
		"add": jsonrpc.NewHandler(
			func(ctx context.Context, request interface{}) (interface{}, error) {
				return nil, nil
			},
			func(context.Context, json.RawMessage) (request interface{}, err error) {
				return nil, nil
			},
			func(_ context.Context, eResp interface{}) (json.RawMessage, error) {
				return json.Marshal(jsonrpc.NewInvalidParamsError("field missing"))
			},
		),
	})

	server := httptest.NewServer(jsonserver)
	defer server.Close()

	resp, err := http.Post(server.URL, "", addBody("add", map[string]int{"a": 1, "b": 3}))

	if err != nil {
		t.Fatalf("Unexpected error '%s'", err)
	}

	buf, _ := ioutil.ReadAll(resp.Body)
	if got, expected := string(buf), `{"jsonrpc":"2.0","result":{"code":-32602,"message":"field missing"},"id":1}`+"\n"; got != expected {
		t.Errorf("Response: expected '%s', actual '%s'", expected, got)
	}
}
