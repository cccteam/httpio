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
func (e *Encoder) encode(body interface{}) error {
	if body == nil {
		body = struct{}{}
	}

	if err := e.encoder.Encode(body); err != nil {
		// If we fail to encode the response, we need to write a 500 status code.
		// This isn't guaranteed to be written if the encoder has already written to the response body,
		// but it will at least catch some cases
		e.w.WriteHeader(http.StatusInternalServerError)

		return errors.WrapSkipFrames(err, "encoder.Encode()", 2)
	}

	return nil
}

// StatusCode writes a statusCode to the response header and returns the original error
func (e *Encoder) StatusCode(statusCode int, err error) error {
	return e.statusCode(statusCode, err)
}

func (e *Encoder) statusCode(statusCode int, err error) error {
	e.w.WriteHeader(statusCode)

	return errors.WrapSkipFrames(err, "handler error", 2)
}

// StatusCodeWithMessage writes a statusCode and message to the response header and returns the original error
func (e *Encoder) StatusCodeWithMessage(statusCode int, err error, message string) error {
	return e.statusCodeWithMessage(statusCode, err, message)
}

// statusCodeWithMessage writes a statusCode and message to the response header and returns the original error
func (e *Encoder) statusCodeWithMessage(statusCode int, err error, message string) error {
	e.w.WriteHeader(statusCode)
	if err := e.encode(&MessageResponse{Message: message}); err != nil {
		return err
	}

	return errors.WrapSkipFrames(err, "handler error", 2)
}

// Ok returns a default http 200 status response with a body
func (e *Encoder) Ok(body interface{}) error {
	return e.ok(body)
}

// ok returns the http 200 status response.
// This ensures the call stack has the same # of frames
func (e *Encoder) ok(body interface{}) error {
	return e.encode(body)
}

// Unauthorized writes a 401 status to the response header and returns the original error
func (e *Encoder) Unauthorized(err error) error {
	return e.statusCode(http.StatusUnauthorized, err)
}

// Forbidden writes a 403 status to the response header and returns the original error
func (e *Encoder) Forbidden(err error) error {
	return e.statusCode(http.StatusForbidden, err)
}

// ForbiddenWithMessage returns a forbidden response with an error message
func (e *Encoder) ForbiddenWithMessage(err error, message string) error {
	return e.statusCodeWithMessage(http.StatusForbidden, err, message)
}

// UnauthorizedWithMessage returns an unauthorized response with an error message
func (e *Encoder) UnauthorizedWithMessage(err error, message string) error {
	return e.statusCodeWithMessage(http.StatusUnauthorized, err, message)
}

// BadRequest writes a http 400 status to the response header and returns the original error
func (e *Encoder) BadRequest(err error) error {
	return e.statusCode(http.StatusBadRequest, err)
}

// BadRequestWithMessage writes a bad request with a message body in the response
func (e *Encoder) BadRequestWithMessage(err error, message string) error {
	return e.statusCodeWithMessage(http.StatusBadRequest, err, message)
}

// InternalServerError writes a 500 status to the response header and returns the original error
func (e *Encoder) InternalServerError(err error) error {
	return e.statusCode(http.StatusInternalServerError, err)
}

// InternalServerErrorWithMessage writes a 500 status to the response header and encodes a message in the response body
func (e *Encoder) InternalServerErrorWithMessage(err error, message string) error {
	return e.statusCodeWithMessage(http.StatusInternalServerError, err, message)
}

// NotFound writes a 404 status to the response header and returns the original error
func (e *Encoder) NotFound(err error) error {
	return e.statusCode(http.StatusNotFound, err)
}

// NotFoundWithMessage writes a 404 status to the response header and encodes a message in the response body
func (e *Encoder) NotFoundWithMessage(err error, message string) error {
	return e.statusCodeWithMessage(http.StatusNotFound, err, message)
}

// Conflict writes a 409 status to the response header and returns the original error
func (e *Encoder) Conflict(err error) error {
	return e.statusCode(http.StatusConflict, err)
}

// ConflictWithMessage writes a 409 status to the response header and encodes a message in the response body
func (e *Encoder) ConflictWithMessage(err error, message string) error {
	return e.statusCodeWithMessage(http.StatusConflict, err, message)
}

// ServiceUnavailable writes a 503 status to the response header and returns the original error
func (e *Encoder) ServiceUnavailable(err error) error {
	return e.statusCode(http.StatusServiceUnavailable, err)
}

// ServiceUnavailableWithMessage writes a 503 status to the response header and encodes a message in the response body
func (e *Encoder) ServiceUnavailableWithMessage(err error, message string) error {
	return e.statusCodeWithMessage(http.StatusServiceUnavailable, err, message)
}
