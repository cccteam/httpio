// httpio handles encoding and decoding for http io. This package is used to standardize request and response handling.
package httpio

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/errors/v5"
	"go.uber.org/mock/gomock"
)

func TestEncoder_encode(t *testing.T) {
	t.Parallel()
	type response struct {
		Message string
	}

	type args struct {
		response interface{}
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		setupEncoder func(encoder *MockHTTPEncoder, w http.ResponseWriter) HTTPEncoder
	}{
		{
			name:    "successfully encodes a response",
			wantErr: false,
			args: args{
				response: &response{
					Message: "this is a good response",
				},
			},
			setupEncoder: func(_ *MockHTTPEncoder, w http.ResponseWriter) HTTPEncoder {
				return json.NewEncoder(w)
			},
		},
		{
			name:    "fails to encode a response",
			wantErr: true,
			args: args{
				response: "Hello world",
			},
			setupEncoder: func(encoder *MockHTTPEncoder, _ http.ResponseWriter) HTTPEncoder {
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
			setupEncoder: func(_ *MockHTTPEncoder, w http.ResponseWriter) HTTPEncoder {
				return json.NewEncoder(w)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockEncoder := NewMockHTTPEncoder(ctrl)

			recorder := httptest.NewRecorder()

			encoder := &Encoder{
				encoder: tt.setupEncoder(mockEncoder, recorder),
				w:       recorder,
			}
			if err := encoder.encode(tt.args.response, 1); (err != nil) != tt.wantErr {
				t.Errorf("Encoder.EncodeResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncoder_statusCodeWithMessage(t *testing.T) {
	t.Parallel()
	type args struct {
		message    string
		err        error
		statusCode int
	}
	tests := []struct {
		name         string
		args         args
		setupEncoder func(e *MockHTTPEncoder)
		wantErr      bool
		wantStatus   int
	}{
		{
			name: "BadRequest message",
			args: args{
				message:    "Testing",
				err:        nil,
				statusCode: http.StatusBadRequest,
			},
			setupEncoder: func(e *MockHTTPEncoder) {
				e.EXPECT().Encode(&MessageResponse{Message: "Testing"}).Return(nil).AnyTimes()
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    false,
		},
		{
			name: "BadRequest message with error",
			args: args{
				message:    "Testing",
				err:        errors.New("some error"),
				statusCode: http.StatusBadRequest,
			},
			setupEncoder: func(e *MockHTTPEncoder) {
				e.EXPECT().Encode(&MessageResponse{Message: "Testing"}).Return(nil).AnyTimes()
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "fails to encode BadRequest message",
			args: args{
				message:    "Testing",
				err:        nil,
				statusCode: http.StatusBadRequest,
			},
			setupEncoder: func(e *MockHTTPEncoder) {
				e.EXPECT().Encode(&MessageResponse{Message: "Testing"}).Return(errors.New("big error")).AnyTimes()
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			e := NewMockHTTPEncoder(ctrl)
			recorder := httptest.NewRecorder()
			tt.setupEncoder(e)
			encoder := &Encoder{
				encoder: e,
				w:       recorder,
			}
			if err := encoder.statusCodeWithMessage(context.Background(), tt.args.statusCode, tt.args.err, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Encoder.Method() error = %v, wantErr %v", err, tt.wantErr)
			}
			if recorder.Result().StatusCode != tt.wantStatus {
				t.Errorf("Encoder.Method() wanted status code %d, got %d", tt.wantStatus, recorder.Result().StatusCode)
			}
		})
	}
}

func TestEncoder_withBody(t *testing.T) {
	t.Parallel()

	type args struct {
		statusCode int
		message    string
	}
	tests := []struct {
		name         string
		args         args
		encodeMethod func(e *Encoder, statusCode int, body interface{}) error
		setupEncoder func(e *MockHTTPEncoder, w http.ResponseWriter) HTTPEncoder
		wantErr      bool
		wantStatus   int
	}{
		{
			name: "Ok",
			args: args{
				statusCode: http.StatusOK,
				message:    "Testing",
			},
			setupEncoder: func(e *MockHTTPEncoder, _ http.ResponseWriter) HTTPEncoder {
				e.EXPECT().Encode("Testing").Return(nil).AnyTimes()
				return e
			},
			encodeMethod: func(e *Encoder, _ int, body interface{}) error {
				return e.Ok(body)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "Ok with error",
			args: args{
				statusCode: http.StatusInternalServerError,
				message:    "Testing",
			},
			setupEncoder: func(e *MockHTTPEncoder, _ http.ResponseWriter) HTTPEncoder {
				e.EXPECT().Encode("Testing").Return(errors.New("big error")).AnyTimes()
				return e
			},
			encodeMethod: func(e *Encoder, _ int, body interface{}) error {
				return e.Ok(body)
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name: "StatusCodeWithBody",
			args: args{
				statusCode: http.StatusBadRequest,
				message:    "Testing",
			},
			setupEncoder: func(e *MockHTTPEncoder, _ http.ResponseWriter) HTTPEncoder {
				e.EXPECT().Encode("Testing").Return(nil).AnyTimes()
				return e
			},
			encodeMethod: func(e *Encoder, statusCode int, body interface{}) error {
				return e.StatusCodeWithBody(statusCode, body)
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			e := NewMockHTTPEncoder(ctrl)
			recorder := httptest.NewRecorder()
			encoder := &Encoder{
				encoder: tt.setupEncoder(e, recorder),
				w:       recorder,
			}
			if err := tt.encodeMethod(encoder, tt.args.statusCode, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Encoder.Method() error = %v, wantErr %v", err, tt.wantErr)
			}

			if recorder.Result().StatusCode != tt.wantStatus {
				t.Errorf("Wanted response status code %d, got %d", tt.wantStatus, recorder.Result().StatusCode)
			}
		})
	}
}

func TestEncoder_encodeMethods(t *testing.T) {
	t.Parallel()

	type args struct {
		message string
		a       []interface{}
		err     error
	}
	tests := []struct {
		name              string
		e                 *Encoder
		args              args
		encodeMethod      func(e *Encoder, msg string, a []interface{}, err error) error
		wantStatus        int
		wantMessage       string
		wantErr           bool
		wantContainsError bool
	}{
		{
			name: "BadRequest()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.BadRequest(context.Background())
			},
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "BadRequestWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.BadRequestWithError(context.Background(), err)
			},
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "BadRequestMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.BadRequestMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "BadRequestMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.BadRequestMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "BadRequestMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.BadRequestMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "BadRequestMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.BadRequestMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "Unauthorized()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.Unauthorized(context.Background())
			},
			wantStatus:        http.StatusUnauthorized,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "UnauthorizedWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.UnauthorizedWithError(context.Background(), err)
			},
			wantStatus:        http.StatusUnauthorized,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "UnauthorizedMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.UnauthorizedMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusUnauthorized,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "UnauthorizedMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.UnauthorizedMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusUnauthorized,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "UnauthorizedMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.UnauthorizedMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusUnauthorized,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "UnauthorizedMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.UnauthorizedMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusUnauthorized,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "Forbidden()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.Forbidden(context.Background())
			},
			wantStatus:        http.StatusForbidden,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "ForbiddenWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.ForbiddenWithError(context.Background(), err)
			},
			wantStatus:        http.StatusForbidden,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "ForbiddenMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.ForbiddenMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusForbidden,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "ForbiddenMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.ForbiddenMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusForbidden,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "ForbiddenMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.ForbiddenMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusForbidden,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "ForbiddenMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.ForbiddenMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusForbidden,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "NotFound()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.NotFound(context.Background())
			},
			wantStatus:        http.StatusNotFound,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "NotFoundWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.NotFoundWithError(context.Background(), err)
			},
			wantStatus:        http.StatusNotFound,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "NotFoundMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.NotFoundMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusNotFound,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "NotFoundMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.NotFoundMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusNotFound,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "NotFoundMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.NotFoundMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusNotFound,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "NotFoundMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.NotFoundMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusNotFound,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "MethodNotAllowed()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.MethodNotAllowed(context.Background())
			},
			wantStatus:        http.StatusMethodNotAllowed,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "MethodNotAllowedWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.MethodNotAllowedWithError(context.Background(), err)
			},
			wantStatus:        http.StatusMethodNotAllowed,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "MethodNotAllowedMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.MethodNotAllowedMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusMethodNotAllowed,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "MethodNotAllowedMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.MethodNotAllowedMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusMethodNotAllowed,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "MethodNotAllowedMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.MethodNotAllowedMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusMethodNotAllowed,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "MethodNotAllowedMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.MethodNotAllowedMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusMethodNotAllowed,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "NotAcceptable()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.NotAcceptable(context.Background())
			},
			wantStatus:        http.StatusNotAcceptable,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "NotAcceptableWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.NotAcceptableWithError(context.Background(), err)
			},
			wantStatus:        http.StatusNotAcceptable,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "NotAcceptableMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.NotAcceptableMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusNotAcceptable,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "NotAcceptableMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.NotAcceptableMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusNotAcceptable,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "NotAcceptableMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.NotAcceptableMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusNotAcceptable,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "NotAcceptableMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.NotAcceptableMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusNotAcceptable,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "RequestTimeout()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.RequestTimeout(context.Background())
			},
			wantStatus:        http.StatusRequestTimeout,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "RequestTimeoutWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.RequestTimeoutWithError(context.Background(), err)
			},
			wantStatus:        http.StatusRequestTimeout,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "RequestTimeoutMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.RequestTimeoutMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusRequestTimeout,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "RequestTimeoutMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.RequestTimeoutMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusRequestTimeout,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "RequestTimeoutMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.RequestTimeoutMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusRequestTimeout,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "RequestTimeoutMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.RequestTimeoutMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusRequestTimeout,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "Conflict()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.Conflict(context.Background())
			},
			wantStatus:        http.StatusConflict,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "ConflictWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.ConflictWithError(context.Background(), err)
			},
			wantStatus:        http.StatusConflict,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "ConflictMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.ConflictMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusConflict,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "ConflictMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.ConflictMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusConflict,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "ConflictMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.ConflictMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusConflict,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "ConflictMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.ConflictMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusConflict,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "RequestEntityTooLarge()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.RequestEntityTooLarge(context.Background())
			},
			wantStatus:        http.StatusRequestEntityTooLarge,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "RequestEntityTooLargeWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.RequestEntityTooLargeWithError(context.Background(), err)
			},
			wantStatus:        http.StatusRequestEntityTooLarge,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "RequestEntityTooLargeMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.RequestEntityTooLargeMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusRequestEntityTooLarge,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "RequestEntityTooLargeMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.RequestEntityTooLargeMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusRequestEntityTooLarge,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "RequestEntityTooLargeMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.RequestEntityTooLargeMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusRequestEntityTooLarge,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "RequestEntityTooLargeMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.RequestEntityTooLargeMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusRequestEntityTooLarge,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "UnsupportedMediaType()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.UnsupportedMediaType(context.Background())
			},
			wantStatus:        http.StatusUnsupportedMediaType,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "UnsupportedMediaTypeWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.UnsupportedMediaTypeWithError(context.Background(), err)
			},
			wantStatus:        http.StatusUnsupportedMediaType,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "UnsupportedMediaTypeMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.UnsupportedMediaTypeMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusUnsupportedMediaType,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "UnsupportedMediaTypeMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.UnsupportedMediaTypeMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusUnsupportedMediaType,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "UnsupportedMediaTypeMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.UnsupportedMediaTypeMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusUnsupportedMediaType,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "UnsupportedMediaTypeMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.UnsupportedMediaTypeMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusUnsupportedMediaType,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "UnprocessableEntity()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.UnprocessableEntity(context.Background())
			},
			wantStatus:        http.StatusUnprocessableEntity,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "UnprocessableEntityWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.UnprocessableEntityWithError(context.Background(), err)
			},
			wantStatus:        http.StatusUnprocessableEntity,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "UnprocessableEntityMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.UnprocessableEntityMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusUnprocessableEntity,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "UnprocessableEntityMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.UnprocessableEntityMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusUnprocessableEntity,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "UnprocessableEntityMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.UnprocessableEntityMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusUnprocessableEntity,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "UnprocessableEntityMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.UnprocessableEntityMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusUnprocessableEntity,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "TooManyRequests()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.TooManyRequests(context.Background())
			},
			wantStatus:        http.StatusTooManyRequests,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "TooManyRequestsWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.TooManyRequestsWithError(context.Background(), err)
			},
			wantStatus:        http.StatusTooManyRequests,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "TooManyRequestsMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.TooManyRequestsMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusTooManyRequests,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "TooManyRequestsMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.TooManyRequestsMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusTooManyRequests,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "TooManyRequestsMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.TooManyRequestsMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusTooManyRequests,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "TooManyRequestsMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.TooManyRequestsMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusTooManyRequests,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "ClientClosedRequest()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.ClientClosedRequest(context.Background())
			},
			wantStatus:        499,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "ClientClosedRequestWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.ClientClosedRequestWithError(context.Background(), err)
			},
			wantStatus:        499,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "ClientClosedRequestMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.ClientClosedRequestMessage(context.Background(), msg)
			},
			wantStatus:        499,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "ClientClosedRequestMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.ClientClosedRequestMessagef(context.Background(), msg, a...)
			},
			wantStatus:        499,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "ClientClosedRequestMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.ClientClosedRequestMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        499,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "ClientClosedRequestMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.ClientClosedRequestMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        499,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "InternalServerError()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.InternalServerError(context.Background())
			},
			wantStatus:        http.StatusInternalServerError,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "InternalServerErrorWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.InternalServerErrorWithError(context.Background(), err)
			},
			wantStatus:        http.StatusInternalServerError,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "InternalServerErrorMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.InternalServerErrorMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusInternalServerError,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "InternalServerErrorMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.InternalServerErrorMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusInternalServerError,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "InternalServerErrorMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.InternalServerErrorMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusInternalServerError,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "InternalServerErrorMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.InternalServerErrorMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusInternalServerError,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "NotImplemented()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.NotImplemented(context.Background())
			},
			wantStatus:        http.StatusNotImplemented,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "NotImplementedWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.NotImplementedWithError(context.Background(), err)
			},
			wantStatus:        http.StatusNotImplemented,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "NotImplementedMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.NotImplementedMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusNotImplemented,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "NotImplementedMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.NotImplementedMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusNotImplemented,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "NotImplementedMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.NotImplementedMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusNotImplemented,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "NotImplementedMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.NotImplementedMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusNotImplemented,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "BadGateway()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.BadGateway(context.Background())
			},
			wantStatus:        http.StatusBadGateway,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "BadGatewayWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.BadGatewayWithError(context.Background(), err)
			},
			wantStatus:        http.StatusBadGateway,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "BadGatewayMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.BadGatewayMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusBadGateway,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "BadGatewayMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.BadGatewayMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusBadGateway,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "BadGatewayMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.BadGatewayMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusBadGateway,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "BadGatewayMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.BadGatewayMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusBadGateway,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "ServiceUnavailable()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.ServiceUnavailable(context.Background())
			},
			wantStatus:        http.StatusServiceUnavailable,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "ServiceUnavailableWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.ServiceUnavailableWithError(context.Background(), err)
			},
			wantStatus:        http.StatusServiceUnavailable,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "ServiceUnavailableMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.ServiceUnavailableMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusServiceUnavailable,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "ServiceUnavailableMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.ServiceUnavailableMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusServiceUnavailable,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "ServiceUnavailableMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.ServiceUnavailableMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusServiceUnavailable,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "ServiceUnavailableMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.ServiceUnavailableMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusServiceUnavailable,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "GatewayTimeout()",
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, _ error) error {
				return e.GatewayTimeout(context.Background())
			},
			wantStatus:        http.StatusGatewayTimeout,
			wantMessage:       "",
			wantErr:           false,
			wantContainsError: false,
		},
		{
			name: "GatewayTimeoutWithError()",
			args: args{
				err: errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, _ string, _ []interface{}, err error) error {
				return e.GatewayTimeoutWithError(context.Background(), err)
			},
			wantStatus:        http.StatusGatewayTimeout,
			wantMessage:       "",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "GatewayTimeoutMessage()",
			args: args{
				message: "Testing",
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, _ error) error {
				return e.GatewayTimeoutMessage(context.Background(), msg)
			},
			wantStatus:        http.StatusGatewayTimeout,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "GatewayTimeoutMessagef",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, _ error) error {
				return e.GatewayTimeoutMessagef(context.Background(), msg, a...)
			},
			wantStatus:        http.StatusGatewayTimeout,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: false,
		},
		{
			name: "GatewayTimeoutMessageWithError()",
			args: args{
				message: "Testing",
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, _ []interface{}, err error) error {
				return e.GatewayTimeoutMessageWithError(context.Background(), err, msg)
			},
			wantStatus:        http.StatusGatewayTimeout,
			wantMessage:       "Testing",
			wantErr:           true,
			wantContainsError: true,
		},
		{
			name: "GatewayTimeoutMessageWithErrorf",
			args: args{
				message: "Testing %s",
				a:       []interface{}{"f"},
				err:     errors.New("Testing"),
			},
			encodeMethod: func(e *Encoder, msg string, a []interface{}, err error) error {
				return e.GatewayTimeoutMessageWithErrorf(context.Background(), err, msg, a...)
			},
			wantStatus:        http.StatusGatewayTimeout,
			wantMessage:       "Testing f",
			wantErr:           true,
			wantContainsError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			encoder := NewEncoder(recorder)
			err := tt.encodeMethod(encoder, tt.args.message, tt.args.a, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encoder.Method() error = %v, wantErr %v", err, tt.wantErr)
			}
			if CauseIsError(err) != tt.wantContainsError {
				t.Errorf("CauseIsError() = %v, wantContainsError %v", err, tt.wantContainsError)
			}

			if recorder.Result().StatusCode != tt.wantStatus {
				t.Errorf("Wanted response status code %d, got %d", tt.wantStatus, recorder.Result().StatusCode)
			}

			if tt.wantMessage == "" {
				if recorder.Body.Len() != 0 {
					t.Errorf("Expected empty body, got %s", recorder.Body.String())
				}

				return
			}

			var bod MessageResponse
			if err := json.NewDecoder(recorder.Result().Body).Decode(&bod); err != nil {
				t.Fatal("failed to decode body")
			}

			if bod.Message != tt.wantMessage {
				t.Errorf("Encoder.statusCodeWithMessage() want message = %s, got = %s", tt.wantMessage, bod.Message)
			}
		})
	}
}

func TestEncoder_ClientMessage(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name        string
		e           *Encoder
		args        args
		wantMessage string
		wantStatus  int
	}{
		{
			name: "BadRequest",
			args: args{
				err: NewBadRequestMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "Unauthorized",
			args: args{
				err: NewUnauthorizedMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name: "Forbidden",
			args: args{
				err: NewForbiddenMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusForbidden,
		},
		{
			name: "NotFound",
			args: args{
				err: NewNotFoundMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusNotFound,
		},
		{
			name: "MethodNotAllowed",
			args: args{
				err: NewMethodNotAllowedMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusMethodNotAllowed,
		},
		{
			name: "NotAcceptable",
			args: args{
				err: NewNotAcceptableMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusNotAcceptable,
		},
		{
			name: "RequestTimeout",
			args: args{
				err: NewRequestTimeoutMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusRequestTimeout,
		},
		{
			name: "Conflict",
			args: args{
				err: NewConflictMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusConflict,
		},
		{
			name: "RequestEntityTooLarge",
			args: args{
				err: NewRequestEntityTooLargeMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusRequestEntityTooLarge,
		},
		{
			name: "UnsupportedMediaType",
			args: args{
				err: NewUnsupportedMediaTypeMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusUnsupportedMediaType,
		},
		{
			name: "UnprocessableEntity",
			args: args{
				err: NewUnprocessableEntityMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			name: "TooManyRequests",
			args: args{
				err: NewTooManyRequestsMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusTooManyRequests,
		},
		{
			name: "ClientClosedRequest",
			args: args{
				err: NewClientClosedRequestMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  499,
		},
		{
			name: "InternalServerError",
			args: args{
				err: NewInternalServerErrorMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name: "NotImplemented",
			args: args{
				err: NewNotImplementedMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusNotImplemented,
		},
		{
			name: "BadGateway",
			args: args{
				err: NewBadGatewayMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusBadGateway,
		},
		{
			name: "ServiceUnavailable",
			args: args{
				err: NewServiceUnavailableMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusServiceUnavailable,
		},
		{
			name: "GatewayTimeout",
			args: args{
				err: NewGatewayTimeoutMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusGatewayTimeout,
		},
		{
			name: "Other Error",
			args: args{
				err: errors.New("Testing"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			encoder := NewEncoder(recorder)
			if err := encoder.ClientMessage(context.Background(), tt.args.err); err == nil {
				t.Errorf("Encoder.ClientMessage() error = %v, wantErr %v", err, true)
			}

			if recorder.Result().StatusCode != tt.wantStatus {
				t.Errorf("Wanted response status code %d, got %d", tt.wantStatus, recorder.Result().StatusCode)
			}

			if tt.wantMessage == "" {
				return
			}

			var bod MessageResponse
			if err := json.NewDecoder(recorder.Result().Body).Decode(&bod); err != nil {
				t.Fatal("failed to decode body")
			}

			if bod.Message != tt.wantMessage {
				t.Errorf("Encoder.ClientMessage() want message = %s, got = %s", tt.wantMessage, bod.Message)
			}
		})
	}
}
