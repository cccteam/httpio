// httpio handles encoding and decoding for http io. This package is used to standardize request and response handling.
package httpio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/errors/v5"
	"github.com/golang/mock/gomock"
)

func TestEncoder_EncodeResponse(t *testing.T) {
	t.Parallel()
	type response struct {
		Message string
	}
	ctrl := gomock.NewController(t)

	type args struct {
		response interface{}
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		setupEncoder func(w http.ResponseWriter) HTTPEncoder
	}{
		{
			name:    "successfully encodes a response",
			wantErr: false,
			args: args{
				response: &response{
					Message: "this is a good response",
				},
			},
			setupEncoder: func(w http.ResponseWriter) HTTPEncoder {
				return json.NewEncoder(w)
			},
		},
		{
			name:    "fails to encode a response",
			wantErr: true,
			args: args{
				response: "Hello world",
			},
			setupEncoder: func(w http.ResponseWriter) HTTPEncoder {
				encoder := NewMockHTTPEncoder(ctrl)
				encoder.EXPECT().Encode("Hello world").Return(errors.New("Failed to encode"))

				return encoder
			},
		},
		{
			name:    "empty response body",
			wantErr: false,
			args: args{
				response: nil,
			},
			setupEncoder: func(w http.ResponseWriter) HTTPEncoder {
				return json.NewEncoder(w)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()

			encoder := &Encoder{
				encoder: tt.setupEncoder(recorder),
				w:       recorder,
			}
			if err := encoder.encode(tt.args.response); (err != nil) != tt.wantErr {
				t.Errorf("Encoder.EncodeResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncoder_StatusCodes(t *testing.T) {
	t.Parallel()
	err := errors.New("This is a test")
	type args struct {
		err error
	}
	tests := []struct {
		name       string
		e          *Encoder
		args       args
		err        error
		statusFunc func(e *Encoder, err error) error
		wantStatus int
	}{
		{
			name: "Successfully writes unauthorized",
			args: args{
				err: err,
			},
			err: err,
			statusFunc: func(e *Encoder, err error) error {
				return e.Unauthorized(err)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "Successfully writes bad request",
			args: args{
				err: err,
			},
			err: err,
			statusFunc: func(e *Encoder, err error) error {
				return e.BadRequest(err)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Successfully writes internal server error",
			args: args{
				err: err,
			},
			err: err,
			statusFunc: func(e *Encoder, err error) error {
				return e.InternalServerError(err)
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "Successfully writes NotFound",
			args: args{
				err: err,
			},
			err: err,
			statusFunc: func(e *Encoder, err error) error {
				return e.NotFound(err)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "Successfully writes Conflict",
			args: args{
				err: err,
			},
			err: err,
			statusFunc: func(e *Encoder, err error) error {
				return e.Conflict(err)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "Successfully writes ServiceUnavailable",
			args: args{
				err: err,
			},
			err: err,
			statusFunc: func(e *Encoder, err error) error {
				return e.ServiceUnavailable(err)
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "Successfully writes a specific status code",
			args: args{
				err: err,
			},
			err: err,
			statusFunc: func(e *Encoder, err error) error {
				return e.StatusCode(122, err)
			},
			wantStatus: 122,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			encoder := NewEncoder(recorder)

			if err := tt.statusFunc(encoder, tt.args.err); !errors.Is(errors.Cause(err), errors.Cause(tt.err)) {
				t.Errorf("Encoder error = %v, want %v", err, tt.err)
			}

			if recorder.Result().StatusCode != tt.wantStatus {
				t.Errorf("wanted status code %d, got %d", tt.wantStatus, recorder.Result().StatusCode)
			}
		})
	}
}

func TestEncoder_StatusCodeWithMessage(t *testing.T) {
	t.Parallel()

	msg := "Testing"
	type args struct {
		message string
		err     error
	}
	tests := []struct {
		name                  string
		e                     *Encoder
		args                  args
		wantErr               bool
		statusWithMessageFunc func(e *Encoder, msg string, err error) error
		wantStatus            int
	}{
		{
			name: "successfully writes a response with Unauthorized status",
			args: args{
				message: msg,
				err:     errors.New("This is a test"),
			},
			wantErr: true,
			statusWithMessageFunc: func(e *Encoder, msg string, err error) error {
				return e.UnauthorizedWithMessage(err, msg)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "successfully writes a response with a Internal Server Error status",
			args: args{
				message: msg,
				err:     errors.New("This is a test"),
			},
			wantErr: true,
			statusWithMessageFunc: func(e *Encoder, msg string, err error) error {
				return e.InternalServerErrorWithMessage(err, msg)
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "successfully writes a response with a Bad Request status",
			args: args{
				message: msg,
				err:     errors.New("This is a test"),
			},
			wantErr: true,
			statusWithMessageFunc: func(e *Encoder, msg string, err error) error {
				return e.BadRequestWithMessage(err, msg)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "successfully writes a response with a NotFound status",
			args: args{
				message: msg,
				err:     errors.New("This is a test"),
			},
			wantErr: true,
			statusWithMessageFunc: func(e *Encoder, msg string, err error) error {
				return e.NotFoundWithMessage(err, msg)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "successfully writes a response with a Conflict status",
			args: args{
				message: msg,
				err:     errors.New("This is a test"),
			},
			wantErr: true,
			statusWithMessageFunc: func(e *Encoder, msg string, err error) error {
				return e.ConflictWithMessage(err, msg)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "successfully writes a response with a ServiceUnavailable status",
			args: args{
				message: msg,
				err:     errors.New("This is a test"),
			},
			wantErr: true,
			statusWithMessageFunc: func(e *Encoder, msg string, err error) error {
				return e.ServiceUnavailableWithMessage(err, msg)
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "successfully writes a response with a specific status",
			args: args{
				message: msg,
				err:     errors.New("This is a test"),
			},
			wantErr: true,
			statusWithMessageFunc: func(e *Encoder, msg string, err error) error {
				return e.StatusCodeWithMessage(202, err, msg)
			},
			wantStatus: 202,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			encoder := NewEncoder(recorder)
			if err := tt.statusWithMessageFunc(encoder, tt.args.message, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("Encoder.StatusCodeWithMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if recorder.Result().StatusCode != tt.wantStatus {
				t.Errorf("Decoder.DecodeRequest() wanted status code %d, got %d", http.StatusUnauthorized, recorder.Result().StatusCode)
			}

			var bod MessageResponse
			if err := json.NewDecoder(recorder.Result().Body).Decode(&bod); err != nil {
				t.Fatal("failed to decode body")
			}

			if bod.Message != tt.args.message {
				t.Errorf("Encoder.StatusCodeWithMessage() want message = %s, got = %s", tt.args.message, bod.Message)
			}
		})
	}
}

func TestEncoder_Ok(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	msg := "Testing"
	type args struct {
		message string
		err     error
	}
	tests := []struct {
		name         string
		e            *Encoder
		args         args
		wantErr      bool
		wantStatus   int
		setupEncoder func(w http.ResponseWriter) HTTPEncoder
	}{
		{
			name: "successfully writes a response with 200 status",
			args: args{
				message: msg,
				err:     nil,
			},
			wantErr: false,
			setupEncoder: func(w http.ResponseWriter) HTTPEncoder {
				e := NewMockHTTPEncoder(ctrl)
				e.EXPECT().Encode(msg).Return(nil).AnyTimes()
				return e
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "fails to write a response with 200 status",
			args: args{
				message: msg,
				err:     nil,
			},
			wantErr: true,
			setupEncoder: func(w http.ResponseWriter) HTTPEncoder {
				e := NewMockHTTPEncoder(ctrl)
				e.EXPECT().Encode(msg).Return(errors.New("Hi, I failed")).AnyTimes()
				return e
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			encoder := &Encoder{
				encoder: tt.setupEncoder(recorder),
				w:       recorder,
			}
			if err := encoder.Ok(tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Encoder.Ok() error = %v, wantErr %v", err, tt.wantErr)
			}

			if recorder.Result().StatusCode != tt.wantStatus {
				t.Errorf("Decoder.DecodeRequest() wanted status code %d, got %d", tt.wantStatus, recorder.Result().StatusCode)
			}
		})
	}
}
