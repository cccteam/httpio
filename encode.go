package httpio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cccteam/logger"
	"github.com/go-playground/errors/v5"
)

// MessageResponse holds a standard structure for http responses that carry a single message
type MessageResponse struct {
	Message string `json:"message,omitempty"`
	TraceID string `json:"traceId,omitempty"`
}

// HTTPEncoder is an interface that is accepted when encoding http responses
type HTTPEncoder interface {
	// Encode is the call that is made to encode data into a response body and returns an error if it fails
	Encode(v interface{}) error
}

// Encoder is a struct that is used for encoding http responses
type Encoder struct {
	// w holds the http response writer to encode responses to
	w http.ResponseWriter
	// encoder holds the encoder that will write to the response
	encoder HTTPEncoder
}

// NewEncoder returns a new Encoder to write to the ResponseWriter
// This encoder will write to the ResponseWriter using a json encoder.
func NewEncoder(w http.ResponseWriter) *Encoder {
	w.Header().Set("Content-Type", "application/json")

	return &Encoder{
		encoder: json.NewEncoder(w),
		w:       w,
	}
}

// encode attempts to encode and write to the response writer
func (e *Encoder) encode(body interface{}, skipFrames uint) error {
	if body == nil {
		return nil
	}

	if err := e.encoder.Encode(body); err != nil {
		// If we fail to encode the response, we need to write a 500 status code.
		// This isn't guaranteed to be written if the encoder has already written to the response body,
		// but it will at least catch some cases
		e.w.WriteHeader(http.StatusInternalServerError)

		return errors.WrapSkipFrames(err, "encoder.Encode()", skipFrames)
	}

	return nil
}

// statusCodeWithMessage writes a statusCode and message to the response header and returns the original error
func (e *Encoder) statusCodeWithMessage(ctx context.Context, statusCode int, err error, message string) error {
	e.w.WriteHeader(statusCode)
	if err := e.encode(&MessageResponse{Message: message, TraceID: logger.Ctx(ctx).TraceID()}, 4); err != nil {
		return err
	}

	return err
}

// StatusCodeWithBody writes a statusCode and body
func (e *Encoder) StatusCodeWithBody(statusCode int, body interface{}) error {
	e.w.WriteHeader(statusCode)

	return e.encode(body, 2)
}

// Ok returns a default http 200 status response with a body
func (e *Encoder) Ok(body interface{}) error {
	return e.encode(body, 2)
}

// BadRequest creates a new empty client message with a BadRequest (400) return code
func (e *Encoder) BadRequest(ctx context.Context) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: badRequest,
	}, "")
}

// Unauthorized creates a new empty client message with a Unauthorized (401) return code
func (e *Encoder) Unauthorized(ctx context.Context) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: unauthorized,
	}, "")
}

// Forbidden creates a new empty client message with a Forbidden (403) return code
func (e *Encoder) Forbidden(ctx context.Context) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: forbidden,
	}, "")
}

// NotFound creates a new empty client message with a NotFound (404) return code
func (e *Encoder) NotFound(ctx context.Context) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: notFound,
	}, "")
}

// Conflict creates a new empty client message with a Conflict (409) return code
func (e *Encoder) Conflict(ctx context.Context) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: conflict,
	}, "")
}

// InternalServerError creates a new empty client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerError(ctx context.Context) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: internalServerError,
	}, "")
}

// ServiceUnavailable creates a new empty client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailable(ctx context.Context) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: serviceUnavailable,
	}, "")
}

// BadRequestWithError creates a new empty client message with error and a BadRequest (400) return code
func (e *Encoder) BadRequestWithError(ctx context.Context, err error) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: badRequest,
		error:   err,
	}, "")
}

// UnauthorizedWithError creates a new empty client message with error and a Unauthorized (401) return code
func (e *Encoder) UnauthorizedWithError(ctx context.Context, err error) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: unauthorized,
		error:   err,
	}, "")
}

// ForbiddenWithError creates a new empty client message with error and a Forbidden (403) return code
func (e *Encoder) ForbiddenWithError(ctx context.Context, err error) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: forbidden,
		error:   err,
	}, "")
}

// NotFoundWithError creates a new empty client message with error and a NotFound (404) return code
func (e *Encoder) NotFoundWithError(ctx context.Context, err error) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: notFound,
		error:   err,
	}, "")
}

// ConflictWithError creates a new empty client message with error and a Conflict (409) return code
func (e *Encoder) ConflictWithError(ctx context.Context, err error) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: conflict,
		error:   err,
	}, "")
}

// InternalServerErrorWithError creates a new empty client message with error and a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorWithError(ctx context.Context, err error) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: internalServerError,
		error:   err,
	}, "")
}

// ServiceUnavailableWithError creates a new empty client message with error and a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableWithError(ctx context.Context, err error) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType: serviceUnavailable,
		error:   err,
	}, "")
}

// BadRequestMessage creates a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessage(ctx context.Context, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       badRequest,
		clientMessage: message,
	}, "")
}

// UnauthorizedMessage creates a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessage(ctx context.Context, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       unauthorized,
		clientMessage: message,
	}, "")
}

// ForbiddenMessage creates a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessage(ctx context.Context, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       forbidden,
		clientMessage: message,
	}, "")
}

// NotFoundMessage creates a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessage(ctx context.Context, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       notFound,
		clientMessage: message,
	}, "")
}

// ConflictMessage creates a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessage(ctx context.Context, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       conflict,
		clientMessage: message,
	}, "")
}

// InternalServerErrorMessage creates a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessage(ctx context.Context, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       internalServerError,
		clientMessage: message,
	}, "")
}

// ServiceUnavailableMessage creates a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessage(ctx context.Context, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: message,
	}, "")
}

// BadRequestMessagef creates a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessagef(ctx context.Context, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       badRequest,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// UnauthorizedMessagef creates a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessagef(ctx context.Context, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       unauthorized,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// ForbiddenMessagef creates a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessagef(ctx context.Context, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       forbidden,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// NotFoundMessagef creates a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessagef(ctx context.Context, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       notFound,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// ConflictMessagef creates a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessagef(ctx context.Context, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       conflict,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// InternalServerErrorMessagef creates a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessagef(ctx context.Context, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       internalServerError,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// ServiceUnavailableMessagef creates a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessagef(ctx context.Context, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// BadRequestMessageWithError wraps an existing error while creating a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessageWithError(ctx context.Context, err error, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       badRequest,
		clientMessage: message,
		error:         err,
	}, "")
}

// UnauthorizedMessageWithError wraps an existing error while creating a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessageWithError(ctx context.Context, err error, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       unauthorized,
		clientMessage: message,
		error:         err,
	}, "")
}

// ForbiddenMessageWithError wraps an existing error while creating a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessageWithError(ctx context.Context, err error, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       forbidden,
		clientMessage: message,
		error:         err,
	}, "")
}

// NotFoundMessageWithError wraps an existing error while creating a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessageWithError(ctx context.Context, err error, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       notFound,
		clientMessage: message,
		error:         err,
	}, "")
}

// ConflictMessageWithError wraps an existing error while creating a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessageWithError(ctx context.Context, err error, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       conflict,
		clientMessage: message,
		error:         err,
	}, "")
}

// InternalServerErrorMessageWithError wraps an existing error while creating a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessageWithError(ctx context.Context, err error, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       internalServerError,
		clientMessage: message,
		error:         err,
	}, "")
}

// ServiceUnavailableMessageWithError wraps an existing error while creating a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessageWithError(ctx context.Context, err error, message string) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: message,
		error:         err,
	}, "")
}

// BadRequestMessageWithErrorf wraps an existing error while creating a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessageWithErrorf(ctx context.Context, err error, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       badRequest,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// UnauthorizedMessageWithErrorf wraps an existing error while creating a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessageWithErrorf(ctx context.Context, err error, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       unauthorized,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// ForbiddenMessageWithErrorf wraps an existing error while creating a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessageWithErrorf(ctx context.Context, err error, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       forbidden,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// NotFoundMessageWithErrorf wraps an existing error while creating a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessageWithErrorf(ctx context.Context, err error, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       notFound,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// ConflictMessageWithErrorf wraps an existing error while creating a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessageWithErrorf(ctx context.Context, err error, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       conflict,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// InternalServerErrorMessageWithErrorf wraps an existing error while creating a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessageWithErrorf(ctx context.Context, err error, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       internalServerError,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// ServiceUnavailableMessageWithErrorf wraps an existing error while creating a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessageWithErrorf(ctx context.Context, err error, format string, a ...any) error {
	return e.clientMessage(ctx, &ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// ClientMessage sets an http code and formats a client message based upon the
// message type found in the error chain. If no message type is found
// it defaults to InternalServerError (500) with no message
func (e *Encoder) ClientMessage(ctx context.Context, err error) error {
	return e.clientMessage(ctx, err, "handler error")
}

func (e *Encoder) clientMessage(ctx context.Context, err error, prefix string) error {
	var rerr error
	if CauseIsError(err) || Message(err) != "" {
		rerr = errors.WrapSkipFrames(err, prefix, 2)
	}

	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		switch cerr.msgType {
		case badRequest:
			return e.statusCodeWithMessage(ctx, http.StatusBadRequest, rerr, cerr.clientMessage)
		case unauthorized:
			return e.statusCodeWithMessage(ctx, http.StatusUnauthorized, rerr, cerr.clientMessage)
		case forbidden:
			return e.statusCodeWithMessage(ctx, http.StatusForbidden, rerr, cerr.clientMessage)
		case notFound:
			return e.statusCodeWithMessage(ctx, http.StatusNotFound, rerr, cerr.clientMessage)
		case conflict:
			return e.statusCodeWithMessage(ctx, http.StatusConflict, rerr, cerr.clientMessage)
		case internalServerError:
			return e.statusCodeWithMessage(ctx, http.StatusInternalServerError, rerr, cerr.clientMessage)
		case serviceUnavailable:
			return e.statusCodeWithMessage(ctx, http.StatusServiceUnavailable, rerr, cerr.clientMessage)
		}
	}

	return e.statusCodeWithMessage(ctx, http.StatusInternalServerError, rerr, "")
}
