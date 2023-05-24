package httpio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/errors/v5"
)

func TestNewDecoder(t *testing.T) {
	t.Parallel()

	type args struct {
		req           *http.Request
		validatorFunc ValidatorFunc
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Creates a new decoder successfully",
			args: args{
				req:           httptest.NewRequest(http.MethodGet, "/test", strings.NewReader("this is a test")),
				validatorFunc: func(s interface{}) error { return nil },
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewDecoder(tt.args.req, tt.args.validatorFunc)
			if !reflect.DeepEqual(got.request, tt.args.req) {
				t.Errorf("NewDecoder().request = %v, want = %v", got, tt.args.req)
			}

			// Can not compare functions, but you can compare the address of the function to check
			// that the same function was passed through the constructor.
			if fmt.Sprintf("%v", got.validate) != fmt.Sprintf("%v", tt.args.validatorFunc) {
				t.Errorf("NewDecoder().validate = %v, want %v", got.validate, tt.args.validatorFunc)
			}
		})
	}
}

func TestDecoder_Decode(t *testing.T) {
	t.Parallel()

	type request struct {
		Name string
	}
	req := &request{
		Name: "Zach",
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal("failed to marshal request")
	}

	type args struct {
		body          string
		validatorFunc ValidatorFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "successfully decodes the request",
			args: args{
				body: string(body),
				validatorFunc: func(s interface{}) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "Fails on decoding the request",
			args: args{
				body: "this is a bad json req body",
			},
			wantErr: true,
		},
		{
			name: "fails to validate the request",
			args: args{
				body: string(body),
				validatorFunc: func(s interface{}) error {
					return errors.New("Failed to validate the request")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader(tt.args.body))

			decoder := NewDecoder(r, tt.args.validatorFunc)
			var req request
			if err := decoder.Decode(&req); (err != nil) != tt.wantErr {
				t.Errorf("Decoder.DecodeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
