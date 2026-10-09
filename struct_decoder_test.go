package httpio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/errors/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type validateMock struct {
	structFunc        func(s any) error
	structPartialFunc func(s any, fields ...string) error
}

func (v *validateMock) Struct(s any) error {
	return v.structFunc(s)
}

func (v *validateMock) StructPartial(s any, fields ...string) error {
	return v.structPartialFunc(s, fields...)
}

// decodeRequest carries the field shapes the decoder distinguishes: a field without a
// json tag, a tagged field, a pointer and a slice.
type decodeRequest struct {
	Name     string
	UserName string   `json:"userName"`
	Password *string  `json:"password"`
	Users    []string `json:"users"`
}

func TestStructDecoder_Decode(t *testing.T) {
	t.Parallel()

	password := "pw"

	type args struct {
		method      string
		body        string
		limit       int64
		validate    bool
		validateErr error
	}
	tests := []struct {
		name        string
		args        args
		want        *decodeRequest
		wantCode    int
		wantMessage string
		wantWhole   bool
		wantPartial []string
	}{
		{
			name: "valid body",
			args: args{body: `{"Name":"Zach","userName":"zach","password":"pw","users":["a","b"]}`},
			want: &decodeRequest{Name: "Zach", UserName: "zach", Password: &password, Users: []string{"a", "b"}},
		},
		{
			name: "field without a tag matches its name lower-cased",
			args: args{body: `{"name":"Zach"}`},
			want: &decodeRequest{Name: "Zach"},
		},
		{
			name:        "tagged field matches its tag only",
			args:        args{body: `{"USERNAME":"zach"}`},
			wantCode:    http.StatusBadRequest,
			wantMessage: "invalid field in json - USERNAME",
		},
		{
			name:        "unknown field",
			args:        args{body: `{"Name":"Zach","Nickname":"Z"}`},
			wantCode:    http.StatusBadRequest,
			wantMessage: "invalid field in json - Nickname",
		},
		{
			name:        "two keys reaching one field",
			args:        args{body: `{"Name":"a","name":"b"}`},
			wantCode:    http.StatusBadRequest,
			wantMessage: "json field name Name collides with another field name of different case",
		},
		{
			name:        "null into a value field",
			args:        args{body: `{"Name":null}`},
			wantCode:    http.StatusBadRequest,
			wantMessage: "Name cannot be null",
		},
		{
			name:        "null into a slice field",
			args:        args{body: `{"users":null}`},
			wantCode:    http.StatusBadRequest,
			wantMessage: "users cannot be null",
		},
		{
			name: "null into a pointer field",
			args: args{body: `{"password":null}`},
			want: &decodeRequest{},
		},
		{
			name:        "malformed JSON",
			args:        args{body: "this is a bad json req body"},
			wantCode:    http.StatusBadRequest,
			wantMessage: "failed to decode request body",
		},
		{
			name:        "empty body",
			args:        args{body: ""},
			wantCode:    http.StatusBadRequest,
			wantMessage: "failed to decode request body",
		},
		{
			name:        "body that is not an object",
			args:        args{body: `["Zach"]`},
			wantCode:    http.StatusBadRequest,
			wantMessage: "failed to decode request body",
		},
		{
			name:        "null body",
			args:        args{body: `null`},
			wantCode:    http.StatusBadRequest,
			wantMessage: "failed to decode request body",
		},
		{
			name:        "scalar body",
			args:        args{body: `true`},
			wantCode:    http.StatusBadRequest,
			wantMessage: "failed to decode request body",
		},
		{
			name:        "trailing data after the object",
			args:        args{body: `{"Name":"Zach"} {"Name":"Zach"}`},
			wantCode:    http.StatusBadRequest,
			wantMessage: "failed to decode request body",
		},
		{
			name:        "a body over the route's limit answers 413 naming the limit",
			args:        args{body: `{"Name":"Zach"}`, limit: 8},
			wantCode:    http.StatusRequestEntityTooLarge,
			wantMessage: "the request body exceeds the maximum of 8 bytes",
		},
		{
			name:        "a limit in kibibytes is named in kibibytes",
			args:        args{body: `{"Name":"` + strings.Repeat("Z", 2048) + `"}`, limit: 2048},
			wantCode:    http.StatusRequestEntityTooLarge,
			wantMessage: "the request body exceeds the maximum of 2KB",
		},
		{
			name: "a body within the route's limit decodes",
			args: args{body: `{"Name":"Zach"}`, limit: 64},
			want: &decodeRequest{Name: "Zach"},
		},
		{
			name:        "value the field cannot hold",
			args:        args{body: `{"Name":1}`},
			wantCode:    http.StatusBadRequest,
			wantMessage: "failed to unmarshal request body",
		},
		{
			name:      "validator accepts the whole struct",
			args:      args{body: `{"Name":"Zach"}`, validate: true},
			want:      &decodeRequest{Name: "Zach"},
			wantWhole: true,
		},
		{
			name:        "validator rejects",
			args:        args{body: `{"Name":"Zach"}`, validate: true, validateErr: errors.New("Name is too short")},
			wantCode:    http.StatusBadRequest,
			wantMessage: "failed validating the request",
			wantWhole:   true,
		},
		{
			name:        "PATCH validates the fields the body carried, in struct order",
			args:        args{method: http.MethodPatch, body: `{"users":["a"],"Name":"Zach"}`, validate: true},
			want:        &decodeRequest{Name: "Zach", Users: []string{"a"}},
			wantPartial: []string{"Name", "Users"},
		},
		{
			name:        "PATCH rejected by the validator",
			args:        args{method: http.MethodPatch, body: `{"userName":"zach"}`, validate: true, validateErr: errors.New("UserName is taken")},
			wantCode:    http.StatusBadRequest,
			wantMessage: "failed validating the request",
			wantPartial: []string{"UserName"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			decoder, err := NewStructDecoder[decodeRequest]()
			if err != nil {
				t.Fatalf("NewStructDecoder() error = %v", err)
			}

			var gotWhole bool
			var gotPartial []string
			if tt.args.validate {
				decoder = decoder.WithValidator(&validateMock{
					structFunc: func(_ any) error {
						gotWhole = true

						return tt.args.validateErr
					},
					structPartialFunc: func(_ any, fields ...string) error {
						gotPartial = fields

						return tt.args.validateErr
					},
				})
			}

			method := tt.args.method
			if method == "" {
				method = http.MethodPost
			}
			ctx := context.Background()
			r := httptest.NewRequestWithContext(ctx, method, "/test", strings.NewReader(tt.args.body))
			if tt.args.limit > 0 {
				// The limit a route carries, installed by the router ahead of the handler.
				r.Body = http.MaxBytesReader(httptest.NewRecorder(), r.Body, tt.args.limit)
			}

			got, err := decoder.Decode(r)
			if (err != nil) != (tt.wantCode != 0) {
				t.Fatalf("StructDecoder.Decode() error = %v, wantCode %d", err, tt.wantCode)
			}
			if gotWhole != tt.wantWhole {
				t.Errorf("ValidatorFunc.Struct() called = %v, want %v", gotWhole, tt.wantWhole)
			}
			if diff := cmp.Diff(tt.wantPartial, gotPartial, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ValidatorFunc.StructPartial() fields mismatch (-want +got):\n%s", diff)
			}

			if err == nil {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("StructDecoder.Decode() mismatch (-want +got):\n%s", diff)
				}

				return
			}

			// The response a handler writes for the refusal, as session's and access's do.
			rr := httptest.NewRecorder()
			_ = NewEncoder(rr).ClientMessage(ctx, err)
			if rr.Code != tt.wantCode {
				t.Errorf("Encoder.ClientMessage() code = %d, want %d", rr.Code, tt.wantCode)
			}
			if wantBody := `{"message":"` + tt.wantMessage + `"}`; strings.TrimSpace(rr.Body.String()) != wantBody {
				t.Errorf("Encoder.ClientMessage() body = %q, want %q", strings.TrimSpace(rr.Body.String()), wantBody)
			}
			if got := Message(err); got != tt.wantMessage {
				t.Errorf("Message() = %q, want %q", got, tt.wantMessage)
			}
		})
	}
}

func TestNewStructDecoder(t *testing.T) {
	t.Parallel()

	type collidingRequest struct {
		Name string `json:"name"`
		NAME string
	}

	tests := []struct {
		name      string
		construct func() error
		wantErr   bool
	}{
		{
			name: "fields with distinct wire names",
			construct: func() error {
				_, err := NewStructDecoder[decodeRequest]()

				return err
			},
		},
		{
			name: "a lower-cased name reaching a tagged field",
			construct: func() error {
				_, err := NewStructDecoder[collidingRequest]()

				return err
			},
			wantErr: true,
		},
		{
			name: "a type that is not a struct",
			construct: func() error {
				_, err := NewStructDecoder[string]()

				return err
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := tt.construct(); (err != nil) != tt.wantErr {
				t.Errorf("NewStructDecoder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
