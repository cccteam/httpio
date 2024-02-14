package httpio

import (
	"encoding/json"
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

	if err != nil && (Message(err) != "" || CauseIsError(err)) {
		return errors.WrapSkipFrames(err, "handler error", 3)
	}

	return nil
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
	return e.clientMessage(newBadRequest())
}

// Unauthorized creates a new empty client message with a Unauthorized (401) return code
func (e *Encoder) Unauthorized() error {
	return e.clientMessage(newUnauthorized())
}

// Forbidden creates a new empty client message with a Forbidden (403) return code
func (e *Encoder) Forbidden() error {
	return e.clientMessage(newForbidden())
}

// NotFound creates a new empty client message with a NotFound (404) return code
func (e *Encoder) NotFound() error {
	return e.clientMessage(newNotFound())
}

// Conflict creates a new empty client message with a Conflict (409) return code
func (e *Encoder) Conflict() error {
	return e.clientMessage(newConflict())
}

// InternalServerError creates a new empty client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerError() error {
	return e.clientMessage(newInternalServerError())
}

// ServiceUnavailable creates a new empty client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailable() error {
	return e.clientMessage(newServiceUnavailable())
}

// BadRequestWithError creates a new empty client message with error and a BadRequest (400) return code
func (e *Encoder) BadRequestWithError(err error) error {
	return e.clientMessage(newBadRequestWithError(err))
}

// UnauthorizedWithError creates a new empty client message with error and a Unauthorized (401) return code
func (e *Encoder) UnauthorizedWithError(err error) error {
	return e.clientMessage(newUnauthorizedWithError(err))
}

// ForbiddenWithError creates a new empty client message with error and a Forbidden (403) return code
func (e *Encoder) ForbiddenWithError(err error) error {
	return e.clientMessage(newForbiddenWithError(err))
}

// NotFoundWithError creates a new empty client message with error and a NotFound (404) return code
func (e *Encoder) NotFoundWithError(err error) error {
	return e.clientMessage(newNotFoundWithError(err))
}

// ConflictWithError creates a new empty client message with error and a Conflict (409) return code
func (e *Encoder) ConflictWithError(err error) error {
	return e.clientMessage(newConflictWithError(err))
}

// InternalServerErrorWithError creates a new empty client message with error and a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorWithError(err error) error {
	return e.clientMessage(newInternalServerErrorWithError(err))
}

// ServiceUnavailableWithError creates a new empty client message with error and a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableWithError(err error) error {
	return e.clientMessage(newServiceUnavailableWithError(err))
}

// BadRequestMessage creates a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessage(message string) error {
	return e.clientMessage(newBadRequestMessage(message))
}

// UnauthorizedMessage creates a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessage(message string) error {
	return e.clientMessage(newUnauthorizedMessage(message))
}

// ForbiddenMessage creates a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessage(message string) error {
	return e.clientMessage(newForbiddenMessage(message))
}

// NotFoundMessage creates a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessage(message string) error {
	return e.clientMessage(newNotFoundMessage(message))
}

// ConflictMessage creates a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessage(message string) error {
	return e.clientMessage(newConflictMessage(message))
}

// InternalServerErrorMessage creates a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessage(message string) error {
	return e.clientMessage(newInternalServerErrorMessage(message))
}

// ServiceUnavailableMessage creates a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessage(message string) error {
	return e.clientMessage(newServiceUnavailableMessage(message))
}

// BadRequestMessagef creates a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessagef(format string, a ...any) error {
	return e.clientMessage(newBadRequestMessagef(format, a...))
}

// UnauthorizedMessagef creates a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessagef(format string, a ...any) error {
	return e.clientMessage(newUnauthorizedMessagef(format, a...))
}

// ForbiddenMessagef creates a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessagef(format string, a ...any) error {
	return e.clientMessage(newForbiddenMessagef(format, a...))
}

// NotFoundMessagef creates a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessagef(format string, a ...any) error {
	return e.clientMessage(newNotFoundMessagef(format, a...))
}

// ConflictMessagef creates a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessagef(format string, a ...any) error {
	return e.clientMessage(newConflictMessagef(format, a...))
}

// InternalServerErrorMessagef creates a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessagef(format string, a ...any) error {
	return e.clientMessage(newInternalServerErrorMessagef(format, a...))
}

// ServiceUnavailableMessagef creates a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessagef(format string, a ...any) error {
	return e.clientMessage(newServiceUnavailableMessagef(format, a...))
}

// BadRequestMessageWithError wraps an existing error while creating a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessageWithError(err error, message string) error {
	return e.clientMessage(newBadRequestMessageWithError(err, message))
}

// UnauthorizedMessageWithError wraps an existing error while creating a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessageWithError(err error, message string) error {
	return e.clientMessage(newUnauthorizedMessageWithError(err, message))
}

// ForbiddenMessageWithError wraps an existing error while creating a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessageWithError(err error, message string) error {
	return e.clientMessage(newForbiddenMessageWithError(err, message))
}

// NotFoundMessageWithError wraps an existing error while creating a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessageWithError(err error, message string) error {
	return e.clientMessage(newNotFoundMessageWithError(err, message))
}

// ConflictMessageWithError wraps an existing error while creating a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessageWithError(err error, message string) error {
	return e.clientMessage(newConflictMessageWithError(err, message))
}

// InternalServerErrorMessageWithError wraps an existing error while creating a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessageWithError(err error, message string) error {
	return e.clientMessage(newInternalServerErrorMessageWithError(err, message))
}

// ServiceUnavailableMessageWithError wraps an existing error while creating a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessageWithError(err error, message string) error {
	return e.clientMessage(newServiceUnavailableMessageWithError(err, message))
}

// BadRequestMessageWithErrorf wraps an existing error while creating a new client message with a BadRequest (400) return code
func (e *Encoder) BadRequestMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(newBadRequestMessageWithErrorf(err, format, a...))
}

// UnauthorizedMessageWithErrorf wraps an existing error while creating a new client message with a Unauthorized (401) return code
func (e *Encoder) UnauthorizedMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(newUnauthorizedMessageWithErrorf(err, format, a...))
}

// ForbiddenMessageWithErrorf wraps an existing error while creating a new client message with a Forbidden (403) return code
func (e *Encoder) ForbiddenMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(newForbiddenMessageWithErrorf(err, format, a...))
}

// NotFoundMessageWithErrorf wraps an existing error while creating a new client message with a NotFound (404) return code
func (e *Encoder) NotFoundMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(newNotFoundMessageWithErrorf(err, format, a...))
}

// ConflictMessageWithErrorf wraps an existing error while creating a new client message with a Conflict (409) return code
func (e *Encoder) ConflictMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(newConflictMessageWithErrorf(err, format, a...))
}

// InternalServerErrorMessageWithErrorf wraps an existing error while creating a new client message with a InternalServerError (500) return code
func (e *Encoder) InternalServerErrorMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(newInternalServerErrorMessageWithErrorf(err, format, a...))
}

// ServiceUnavailableMessageWithErrorf wraps an existing error while creating a new client message with a ServiceUnavailable (503) return code
func (e *Encoder) ServiceUnavailableMessageWithErrorf(err error, format string, a ...any) error {
	return e.clientMessage(newServiceUnavailableMessageWithErrorf(err, format, a...))
}

// ClientMessage sets an http code and formats a client message based upon the
// message type found in the error chain. If no message type is found
// it defaults to InternalServerError (500) with no message
func (e *Encoder) ClientMessage(err error) error {
	return e.clientMessage(err)
}

func (e *Encoder) clientMessage(err error) error {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		switch cerr.msgType {
		case badRequest:
			return e.statusCodeWithMessage(http.StatusBadRequest, err, cerr.clientMessage)
		case unauthorized:
			return e.statusCodeWithMessage(http.StatusUnauthorized, err, cerr.clientMessage)
		case forbidden:
			return e.statusCodeWithMessage(http.StatusForbidden, err, cerr.clientMessage)
		case notFound:
			return e.statusCodeWithMessage(http.StatusNotFound, err, cerr.clientMessage)
		case conflict:
			return e.statusCodeWithMessage(http.StatusConflict, err, cerr.clientMessage)
		case internalServerError:
			return e.statusCodeWithMessage(http.StatusInternalServerError, err, cerr.clientMessage)
		case serviceUnavailable:
			return e.statusCodeWithMessage(http.StatusServiceUnavailable, err, cerr.clientMessage)
		}
	}

	return e.statusCodeWithMessage(http.StatusInternalServerError, err, "")
}
