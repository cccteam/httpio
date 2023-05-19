package httpio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/errors/v5"
)

func TestNewDecoder(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader("this is a test"))

	type args struct {
		req           *http.Request
		validatorFunc ValidatorFunc
	}
	tests := []struct {
		name string
		args args
		want *Decoder
	}{
		{
			name: "Creates a new decoder successfully",
			args: args{
				req:           r,
				validatorFunc: func(s interface{}) error { return nil },
			},
			want: &Decoder{
				validateFunc: func(s interface{}) error { return nil },
				request:      r,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewDecoder(tt.args.req, tt.args.validatorFunc)
			if !reflect.DeepEqual(got.request, tt.want.request) {
				t.Errorf("NewDecoder().request = %v, want = %v", got, tt.want)
			}

			if !reflect.DeepEqual(got.validateFunc(got.request), tt.want.validateFunc(tt.want.request)) {
				t.Errorf("NewDecoder().validateReturn = %v, want %v", got, tt.want)
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
