package httpio

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/errors/v5"
)

func TestDecoder_Decode(t *testing.T) {
	t.Parallel()

	type args struct {
		body          string
		validatorFunc ValidatorFunc
	}
	tests := []struct {
		name             string
		args             args
		wantDecodeErr    bool
		wantValidatorErr bool
	}{
		{
			name: "successfully decodes the request",
			args: args{
				body: `{"Name":"Zach"}`,
				validatorFunc: func(_ interface{}) error {
					return nil
				},
			},
		},
		{
			name: "Fails on decoding the request",
			args: args{
				body: "this is a bad json req body",
			},
			wantDecodeErr: true,
		},
		{
			name: "fails to validate the request",
			args: args{
				body: `{"Name":"Zach"}`,
				validatorFunc: func(_ interface{}) error {
					return errors.New("Failed to validate the request")
				},
			},
			wantValidatorErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			type request struct {
				Name string
			}

			decoder, err := NewStructDecoder[request]()
			if err != nil {
				t.Fatalf("NewDecoder() error = %v", err)
			}

			r := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader(tt.args.body))
			if _, err := decoder.Decode(r); (err != nil) != tt.wantDecodeErr {
				t.Fatalf("Decoder.DecodeRequest() error = %v, wantErr %v", err, tt.wantDecodeErr)
			}

			if tt.wantDecodeErr {
				return
			}

			decoder = decoder.WithValidator(tt.args.validatorFunc)

			r = httptest.NewRequest(http.MethodGet, "/test", strings.NewReader(tt.args.body))
			if _, err := decoder.Decode(r); (err != nil) != tt.wantValidatorErr {
				t.Errorf("Decoder.DecodeRequest() error = %v, wantErr %v", err, tt.wantValidatorErr)
			}
		})
	}
}

func TestDecoder_Decode_Error(t *testing.T) {
	t.Parallel()

	type args struct {
		body string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "successfully decodes the request",
			args: args{
				body: `{"Name":"Zach"}`,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			type request struct {
				Name string `json:"name"`
				NAME string
			}

			_, err := NewStructDecoder[request]()
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewDecoder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
