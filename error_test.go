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
			t.Fatalf("Expected it implements jsonrpc.Erroer for type %T", err)
		}

		gotCode, wantCode := jerr.ErrorCode(), c.errCode
		if gotCode != wantCode {
			t.Errorf("ErrorCode(): expected %d, actual %d", wantCode, gotCode)
		}

		gotErr, wantErr := jerr.Error(), c.expMessage
		if gotErr != wantErr {
			t.Errorf("Error(): expected %s, actual %s", wantErr, gotErr)
		}

		data, merr := json.Marshal(err)
		if merr != nil {
			t.Fatalf("Unexpected error marshaling JSON: %s", err)
		}

		gotEnc, wantEnc := string(data), fmt.Sprintf(`{"code":%d,"message":"%s"}`, c.errCode, c.expMessage)
		if gotEnc != wantEnc {
			t.Errorf("JSON: expected %s, actual %s", wantEnc, gotEnc)
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

		gotCode, wantCode := jerr.ErrorCode(), c.errCode
		if gotCode != wantCode {
			t.Errorf("ErrorCode(): expected %d, actual %d", wantCode, gotCode)
		}

		gotErr, wantErr := jerr.Error(), c.errMessage
		if gotErr != wantErr {
			t.Errorf("Error(): expected %s, actual %s", wantErr, gotErr)
		}

		data, merr := json.Marshal(err)
		if merr != nil {
			t.Fatalf("Unexpected error marshaling JSON: %s", err)
		}

		gotEnc, wantEnc := string(data), fmt.Sprintf(`{"code":%d,"message":"%s"}`, c.errCode, c.errMessage)
		if gotEnc != wantEnc {
			t.Errorf("JSON: expected %s, actual %s", wantEnc, gotEnc)
		}
	}
}
