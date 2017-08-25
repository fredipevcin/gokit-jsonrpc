package jsonrpc_test

import (
	"encoding/json"
	"fmt"
	"testing"

	jsonrpc "github.com/fredipevcin/gokit-jsonrpc"
)

func TestCanUnMarshalID(t *testing.T) {
	cases := []struct {
		JSON     string
		expType  string
		expValue interface{}
	}{
		{`12345`, "int", 12345},
		{`12345.6`, "float", 12345.6},
		{`"stringaling"`, "string", "stringaling"},
	}

	for _, c := range cases {
		r := jsonrpc.Request{}
		JSON := fmt.Sprintf(`{"id":%s}`, c.JSON)

		var foo interface{}
		_ = json.Unmarshal([]byte(JSON), &foo)

		err := json.Unmarshal([]byte(JSON), &r)
		if err != nil {
			t.Fatalf("Unexpected error unmarshaling JSON into request: %s\n", err)
		}
		id := r.ID

		switch c.expType {
		case "int":
			want := c.expValue.(int)
			got, err := id.Int()
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("'%s' Int(): want %d, got %d.", c.JSON, want, got)
			}

			// Allow an int ID to be interpreted as a float.
			wantf := float32(c.expValue.(int))
			gotf, err := id.Float32()
			if gotf != wantf {
				t.Fatalf("'%s' Int value as Float32(): want %f, got %f.", c.JSON, wantf, gotf)
			}

			_, err = id.String()
			if err == nil {
				t.Fatal("Expected String() to error for int value. Didn't.")
			}
		case "string":
			want := c.expValue.(string)
			got, err := id.String()
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("'%s' String(): want %s, got %s.", c.JSON, want, got)
			}

			_, err = id.Int()
			if err == nil {
				t.Fatal("Expected Int() to error for string value. Didn't.")
			}
			_, err = id.Float32()
			if err == nil {
				t.Fatal("Expected Float32() to error for string value. Didn't.")
			}
		case "float32":
			want := c.expValue.(float32)
			got, err := id.Float32()
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("'%s' Float32(): want %f, got %f.", c.JSON, want, got)
			}

			_, err = id.String()
			if err == nil {
				t.Fatal("Expected String() to error for float value. Didn't.")
			}
			_, err = id.Int()
			if err == nil {
				t.Fatal("Expected Int() to error for float value. Didn't.")
			}
		}
	}
}

func TestCannotUnMarshalIDInvalidValue(t *testing.T) {
	r := jsonrpc.Request{}

	jsonVal := `{"id":true}`
	err := json.Unmarshal([]byte(jsonVal), &r)
	if err != jsonrpc.ErrParsingRequestID {
		t.Fatalf("Expected error unmarshaling JSON id: %s", jsonVal)
	}

	if r.ID == nil {
		t.Fatal("RequestID should not be nil")
	}

	if r.ID.Error() != jsonrpc.ErrParsingRequestID.Error() {
		t.Fatalf("Unxpected error unmarshaling JSON id: %s", r.ID)
	}
}

func TestCanUnmarshalNullID(t *testing.T) {
	r := jsonrpc.Request{}
	JSON := `{"id":null}`
	err := json.Unmarshal([]byte(JSON), &r)
	if err != nil {
		t.Fatalf("Unexpected error unmarshaling JSON into request: %s\n", err)
	}

	if r.ID != nil {
		t.Fatalf("Expected ID to be nil, got %+v.\n", r.ID)
	}
}

func TestMarshalEmptyRequestID(t *testing.T) {
	r := &jsonrpc.RequestID{}

	resp, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Unexpected error marshaling RequestID: %s\n", err)
	}

	if string(resp) != "null" {
		t.Errorf("Expecting 'null', got %s", string(resp))
	}
}

func TestMarshalRequestID(t *testing.T) {
	cases := []string{
		"11234",
		`"foobar"`,
		`123.456`,
	}

	for _, c := range cases {

		r := &jsonrpc.RequestID{}

		err := json.Unmarshal([]byte(c), &r)
		if err != nil {
			t.Fatalf("TC(%s) Unexpected error unmarshaling RequestID: %s\n", c, err)
		}

		resp, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("TC(%s) Unexpected error marshaling RequestID: %s\n", c, err)
		}

		if got, expect := string(resp), c; got != expect {
			t.Errorf("TC(%s) Expecting %s, got %s", c, expect, got)
		}
	}

}

func TestValidRequest(t *testing.T) {
	cases := []string{
		`{"jsonrpc":"2.0","id":1234,"method":"method"}`,
		`{"jsonrpc":"2.0","id":"string","method":"name"}`,
		`{"jsonrpc":"2.0","id":null,"method":"name"}`,
		`{"jsonrpc":"2.0","method":"name"}`,
	}

	for _, c := range cases {
		var err error
		r := jsonrpc.Request{}

		err = r.Validate()
		if err == nil {
			t.Fatalf("TC(%s) Request should not be valid", c)
		}

		err = json.Unmarshal([]byte(c), &r)
		if err != nil {
			t.Fatalf("TC(%s) Unexpected error unmarshaling Request: %s", c, err)
		}
		err = r.Validate()
		if err != nil {
			t.Fatalf("TC(%s) Request is not valid: %s", c, err)
		}
	}
}
func TestInvalidRequest(t *testing.T) {
	cases := []string{
		`{"jsonrpc":"2.0","id":true,"method":"method"}`,
		`{"jsonrpc":"2.0","id":[],"method":"method"}`,
		`{"jsonrpc":"2.0","id":"string"}`,
		`{"jsonrpc":"1.0","id":null,"method":"name"}`,
		`{"jsonrpc":"2.0","id":"string","method":"rpc.internal"}`,
	}

	for _, c := range cases {
		var err error
		r := jsonrpc.Request{}

		err = r.Validate()
		if err == nil {
			t.Fatalf("TC(%s) Request should not be valid", c)
		}

		err = json.Unmarshal([]byte(c), &r)
		if err != nil && err != jsonrpc.ErrParsingRequestID {
			t.Fatalf("TC(%s) Unexpected error unmarshaling Request: %s", c, err)
		}
		err = r.Validate()
		if err == nil {
			t.Fatalf("TC(%s) Request should not be valid", c)
		}
	}
}
