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
		tt := tt
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
		tt := tt
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
		tt := tt
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
	}
	for _, tt := range tests {
		tt := tt
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
			name: "Conflict",
			args: args{
				err: NewConflictMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusConflict,
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
			name: "InternalServerError",
			args: args{
				err: NewInternalServerErrorMessage("Testing"),
			},
			wantMessage: "Testing",
			wantStatus:  http.StatusInternalServerError,
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
			name: "Other Error",
			args: args{
				err: errors.New("Testing"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		tt := tt
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
