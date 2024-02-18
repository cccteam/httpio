package httpio

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/errors/v5"
)

// MessageResponse holds a standard structure for http responses that carry a single message
type MessageResponse struct {
	Message string `json:"message,omitempty"`
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
func (e *Encoder) statusCodeWithMessage(statusCode int, err error, message string) error {
	e.w.WriteHeader(statusCode)
	if message != "" {
		if err := e.encode(&MessageResponse{Message: message}, 4); err != nil {
			return err
		}
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
func (e *Encoder) BadRequest() error {
	return e.clientMessage(&ClientMessage{
		msgType: badRequest,
	}, "")
}

// Unauthorized creates a new empty client message with a Unauthorized (401) return code
func (e *Encoder) Unauthorized() error {
	return e.clientMessage(&ClientMessage{
		msgType: unauthorized,
	}, "")
}

// Forbidden creates a new empty client message with a Forbidden (403) return code
func (e *Encoder) Forbidden() error {
	return e.clientMessage(&ClientMessage{
		msgType: forbidden,
	}, "")
}

// NotFound creates a new empty client message with a NotFound (404) return code
func (e *Encoder) NotFound() error {
	return e.clientMessage(&ClientMessage{
		msgType: notFound,
	}, "")
}

// Conflict creates a new empty client message with a Conflict (409) return code
func (e *Encoder) Conflict() error {
	return e.clientMessage(&ClientMessage{
		msgType: conflict,
	}, "")
}

// InternalServerError creates a new empty client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerError() error {
	return e.clientMessage(&ClientMessage{
		msgType: internalServerError,
	}, "")
}

// ServiceUnavailable creates a new empty client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailable() error {
	return e.clientMessage(&ClientMessage{
		msgType: serviceUnavailable,
	}, "")
}

// BadRequestWithError creates a new empty client message with error and a BadRequest (400) return code
func (e *Encoder) BadRequestWithError(err error) error {
	return e.clientMessage(&ClientMessage{
		msgType: badRequest,
		error:   err,
	}, "")
}

// UnauthorizedWithError creates a new empty client message with error and a Unauthorized (401) return code
func (e *Encoder) UnauthorizedWithError(err error) error {
	return e.clientMessage(&ClientMessage{
		msgType: unauthorized,
		error:   err,
	}, "")
}

// ForbiddenWithError creates a new empty client message with error and a Forbidden (403) return code
func (e *Encoder) ForbiddenWithError(err error) error {
	return e.clientMessage(&ClientMessage{
		msgType: forbidden,
		error:   err,
	}, "")
}

// NotFoundWithError creates a new empty client message with error and a NotFound (404) return code
func (e *Encoder) NotFoundWithError(err error) error {
	return e.clientMessage(&ClientMessage{
		msgType: notFound,
		error:   err,
	}, "")
}

// ConflictWithError creates a new empty client message with error and a Conflict (409) return code
func (e *Encoder) ConflictWithError(err error) error {
	return e.clientMessage(&ClientMessage{
		msgType: conflict,
		error:   err,
	}, "")
}

// InternalServerErrorWithError creates a new empty client message with error and a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorWithError(err error) error {
	return e.clientMessage(&ClientMessage{
		msgType: internalServerError,
		error:   err,
	}, "")
}

// ServiceUnavailableWithError creates a new empty client message with error and a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableWithError(err error) error {
	return e.clientMessage(&ClientMessage{
		msgType: serviceUnavailable,
		error:   err,
	}, "")
}

// BadRequestMessage creates a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessage(message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       badRequest,
		clientMessage: message,
	}, "")
}

// UnauthorizedMessage creates a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessage(message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: message,
	}, "")
}

// ForbiddenMessage creates a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessage(message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       forbidden,
		clientMessage: message,
	}, "")
}

// NotFoundMessage creates a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessage(message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       notFound,
		clientMessage: message,
	}, "")
}

// ConflictMessage creates a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessage(message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       conflict,
		clientMessage: message,
	}, "")
}

// InternalServerErrorMessage creates a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessage(message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: message,
	}, "")
}

// ServiceUnavailableMessage creates a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessage(message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: message,
	}, "")
}

// BadRequestMessagef creates a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessagef(format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       badRequest,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// UnauthorizedMessagef creates a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessagef(format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// ForbiddenMessagef creates a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessagef(format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       forbidden,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// NotFoundMessagef creates a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessagef(format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       notFound,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// ConflictMessagef creates a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessagef(format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       conflict,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// InternalServerErrorMessagef creates a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessagef(format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// ServiceUnavailableMessagef creates a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessagef(format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: fmt.Sprintf(format, a...),
	}, "")
}

// BadRequestMessageWithError wraps an existing error while creating a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessageWithError(err error, message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       badRequest,
		clientMessage: message,
		error:         err,
	}, "")
}

// UnauthorizedMessageWithError wraps an existing error while creating a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessageWithError(err error, message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: message,
		error:         err,
	}, "")
}

// ForbiddenMessageWithError wraps an existing error while creating a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessageWithError(err error, message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       forbidden,
		clientMessage: message,
		error:         err,
	}, "")
}

// NotFoundMessageWithError wraps an existing error while creating a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessageWithError(err error, message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       notFound,
		clientMessage: message,
		error:         err,
	}, "")
}

// ConflictMessageWithError wraps an existing error while creating a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessageWithError(err error, message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       conflict,
		clientMessage: message,
		error:         err,
	}, "")
}

// InternalServerErrorMessageWithError wraps an existing error while creating a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessageWithError(err error, message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: message,
		error:         err,
	}, "")
}

// ServiceUnavailableMessageWithError wraps an existing error while creating a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessageWithError(err error, message string) error {
	return e.clientMessage(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: message,
		error:         err,
	}, "")
}

// BadRequestMessageWithErrorf wraps an existing error while creating a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       badRequest,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// UnauthorizedMessageWithErrorf wraps an existing error while creating a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// ForbiddenMessageWithErrorf wraps an existing error while creating a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       forbidden,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// NotFoundMessageWithErrorf wraps an existing error while creating a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       notFound,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// ConflictMessageWithErrorf wraps an existing error while creating a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       conflict,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// InternalServerErrorMessageWithErrorf wraps an existing error while creating a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// ServiceUnavailableMessageWithErrorf wraps an existing error while creating a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	}, "")
}

// ClientMessage sets an http code and formats a client message based upon the
// message type found in the error chain. If no message type is found
// it defaults to InternalServerError (500) with no message
func (e *Encoder) ClientMessage(err error) error {
	return e.clientMessage(err, "handler error")
}

func (e *Encoder) clientMessage(err error, prefix string) error {
	var rerr error
	if CauseIsError(err) || Message(err) != "" {
		rerr = errors.WrapSkipFrames(err, prefix, 2)
	}

	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		switch cerr.msgType {
		case badRequest:
			return e.statusCodeWithMessage(http.StatusBadRequest, rerr, cerr.clientMessage)
		case unauthorized:
			return e.statusCodeWithMessage(http.StatusUnauthorized, rerr, cerr.clientMessage)
		case forbidden:
			return e.statusCodeWithMessage(http.StatusForbidden, rerr, cerr.clientMessage)
		case notFound:
			return e.statusCodeWithMessage(http.StatusNotFound, rerr, cerr.clientMessage)
		case conflict:
			return e.statusCodeWithMessage(http.StatusConflict, rerr, cerr.clientMessage)
		case internalServerError:
			return e.statusCodeWithMessage(http.StatusInternalServerError, rerr, cerr.clientMessage)
		case serviceUnavailable:
			return e.statusCodeWithMessage(http.StatusServiceUnavailable, rerr, cerr.clientMessage)
		}
	}

	return e.statusCodeWithMessage(http.StatusInternalServerError, rerr, "")
}
