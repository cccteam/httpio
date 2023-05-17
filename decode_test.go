package httpio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/errors/v5"
	"github.com/golang/mock/gomock"
)

func TestNewDecoder(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader("this is a test"))

	ctrl := gomock.NewController(t)
	v := NewMockValidator(ctrl)
	type args struct {
		req       *http.Request
		validator Validator
	}
	tests := []struct {
		name string
		args args
		want *Decoder
	}{
		{
			name: "Creates a new decoder successfully",
			args: args{
				req:       r,
				validator: v,
			},
			want: &Decoder{
				validator: v,
				request:   r,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewDecoder(tt.args.req, tt.args.validator); !reflect.DeepEqual(got.request, tt.want.request) || !reflect.DeepEqual(got.validator, tt.want.validator) {
				t.Errorf("NewDecoder() = %v, want %v", got, tt.want)
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
		body string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		prepare func(v *MockValidator)
	}{
		{
			name: "successfully decodes the request",
			args: args{
				body: string(body),
			},
			wantErr: false,
			prepare: func(v *MockValidator) {
				v.EXPECT().Struct(req).Return(nil).Times(1)
			},
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
			},
			wantErr: true,
			prepare: func(v *MockValidator) {
				v.EXPECT().Struct(req).Return(errors.New("Failed to validate the request")).Times(1)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			v := NewMockValidator(ctrl)
			if tt.prepare != nil {
				tt.prepare(v)
			}
			r := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader(tt.args.body))
			decoder := NewDecoder(r, v)
			var req request
			if err := decoder.Decode(&req); (err != nil) != tt.wantErr {
				t.Errorf("Decoder.DecodeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
