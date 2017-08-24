package jsonrpc_test

import (
	"encoding/json"
	"fmt"
	"testing"

	jsonrpc "github.com/fredipevcin/gokit-jsonrpc"
)

func TestPredefinedErrors(t *testing.T) {

	cases := []struct {
		errCode    int
		expMessage string
	}{
		{jsonrpc.ParseError, "An error occurred on the server while parsing the JSON text"},
		{jsonrpc.InvalidRequestError, "The JSON sent is not a valid Request object"},
		{jsonrpc.MethodNotFoundError, "The method does not exist / is not available"},
		{jsonrpc.InvalidParamsError, "Invalid method parameter(s)"},
		{jsonrpc.InternalError, "Internal JSON-RPC error"},
	}

	for _, c := range cases {
		var err error
		err = jsonrpc.NewError(c.errCode)
		jerr, ok := err.(jsonrpc.Errorer)
		if !ok {
			t.Fatalf("Expected err implements jsonrpc.Erroer for type %T", err)
		}

		if got, expected := jerr.ErrorCode(), c.errCode; got != expected {
			t.Errorf("ErrorCode(): expected %d, actual %d", expected, got)
		}

		if got, expected := jerr.Error(), c.expMessage; got != expected {
			t.Errorf("Error(): expected %s, actual %s", expected, got)
		}

		data, merr := json.Marshal(err)
		if merr != nil {
			t.Fatalf("Unexpected error marshaling JSON: %s", err)
		}

		if got, expected := string(data), fmt.Sprintf(`{"code":%d,"message":"%s"}`, c.errCode, c.expMessage); got != expected {
			t.Errorf("JSON: expected %s, actual %s", expected, got)
		}
	}
}
func TestCustomErrors(t *testing.T) {
	cases := []struct {
		errCode    int
		errMessage string
	}{
		{123, "msg 1"},
		{456, "msg 2"},
	}

	for _, c := range cases {
		var err error
		err = jsonrpc.NewError(c.errCode, c.errMessage)
		jerr, ok := err.(jsonrpc.Errorer)
		if !ok {
			t.Fatalf("Expected it implements jsonrpc.Erroer for type %T", err)
		}

		if got, expected := jerr.ErrorCode(), c.errCode; got != expected {
			t.Errorf("ErrorCode(): expected %d, actual %d", expected, got)
		}

		if got, expected := jerr.Error(), c.errMessage; got != expected {
			t.Errorf("Error(): expected %s, actual %s", expected, got)
		}

		data, merr := json.Marshal(err)
		if merr != nil {
			t.Fatalf("Unexpected error marshaling JSON: %s", err)
		}

		if got, expected := string(data), fmt.Sprintf(`{"code":%d,"message":"%s"}`, c.errCode, c.errMessage); got != expected {
			t.Errorf("JSON: expected %s, actual %s", expected, got)
		}
	}
}
