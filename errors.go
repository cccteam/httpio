package httpio

import (
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/go-playground/errors/v5"
)

type msgType int

const (
	badRequest          msgType = iota // http code 400
	unauthorized                       // http code 401
	forbidden                          // http code 403
	notFound                           // http code 404
	conflict                           // http code 409
	internalServerError                // http code 500
	serviceUnavailable                 // http code 503
)

func init() {
	errors.RegisterErrorFormatFn(errorFormatFn)
}

// errorFormatFn implements a custom format function for the errors package
// to properly unwrap error chains contained inside a ClientMessage.
func errorFormatFn(chain errors.Chain) string {
	var errMsg strings.Builder

	for i, link := range chain {
		if i == 0 {
			if chain, ok := stderrors.Unwrap(link.Err).(errors.Chain); ok {
				errMsg.WriteString(errorFormatFn(chain))
				errMsg.WriteString("\n")
			}
		}
		if i > 0 {
			errMsg.WriteString("\n")
		}
		errMsg.WriteString(link.Error())
	}

	return errMsg.String()
}

// Message returns the message from the wrapped ClientMessage or an empty string
func Message(err error) string {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		return cerr.Message()
	}

	return ""
}

// Messages returns a slice of messages from the ClientMessage's contained within the chain of errors
func Messages(err error) []string {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		subMsgs := Messages(cerr.Unwrap())

		msgs := make([]string, 0, len(subMsgs)+1)
		msgs = append(msgs, cerr.Message())
		msgs = append(msgs, subMsgs...)

		return msgs
	}

	return nil
}

// CauseIsError returns true if the Cause of this error is an error vs a ClientMessage with nil error
func CauseIsError(err error) bool {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		return CauseIsError(cerr.Unwrap())
	}

	return err != nil
}

// ClientMessage is a custom message type that can be used to return client messages
type ClientMessage struct {
	msgType       msgType
	clientMessage string
	error         error
}

// Message returns the client message
func (c *ClientMessage) Message() string {
	return c.clientMessage
}

// Error returns the error message
func (c *ClientMessage) Error() string {
	if c.error == nil && c.clientMessage == "" {
		return ""
	}
	if c.clientMessage == "" {
		return c.error.Error()
	}

	msg := fmt.Sprintf("Client Message:%q", c.clientMessage)
	if c.error == nil {
		return msg
	}

	return fmt.Sprintf("%s: %s", msg, c.error.Error())
}

func (c *ClientMessage) Unwrap() error {
	return c.error
}

func wrap(err error) errors.Chain {
	return errors.WrapSkipFrames(err, "", 2)
}

// NewBadRequest creates a new empty client message with a BadRequest (400) return code
func NewBadRequest() errors.Chain {
	return wrap(&ClientMessage{
		msgType: badRequest,
	})
}

// NewUnauthorized creates a new empty client message with a Unauthorized (401) return code
func NewUnauthorized() errors.Chain {
	return wrap(&ClientMessage{
		msgType: unauthorized,
	})
}

// NewForbidden creates a new empty client message with a Forbidden (403) return code
func NewForbidden() errors.Chain {
	return wrap(&ClientMessage{
		msgType: forbidden,
	})
}

// NewNotFound creates a new empty client message with a NotFound (404) return code
func NewNotFound() errors.Chain {
	return wrap(&ClientMessage{
		msgType: notFound,
	})
}

// NewConflict creates a new empty client message with a Conflict (409) return code
func NewConflict() errors.Chain {
	return wrap(&ClientMessage{
		msgType: conflict,
	})
}

// NewInternalServerError creates a new empty client message with a InternalServerError (500) return code
func NewInternalServerError() errors.Chain {
	return wrap(&ClientMessage{
		msgType: internalServerError,
	})
}

// NewServiceUnavailable creates a new empty client message with a ServiceUnavailable (503) return code
func NewServiceUnavailable() errors.Chain {
	return wrap(&ClientMessage{
		msgType: serviceUnavailable,
	})
}

// NewBadRequestWithError wraps an existing error while creating a new empty client message and a BadRequest (400) return code
func NewBadRequestWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: badRequest,
		error:   err,
	})
}

// NewUnauthorizedWithError wraps an existing error while creating a new empty client message and a Unauthorized (401) return code
func NewUnauthorizedWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: unauthorized,
		error:   err,
	})
}

// NewForbiddenWithError wraps an existing error while creating a new empty client message and a Forbidden (403) return code
func NewForbiddenWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: forbidden,
		error:   err,
	})
}

// NewNotFoundWithError wraps an existing error while creating a new empty client message and a NotFound (404) return code
func NewNotFoundWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: notFound,
		error:   err,
	})
}

// NewConflictWithError wraps an existing error while creating a new empty client message and a Conflict (409) return code
func NewConflictWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: conflict,
		error:   err,
	})
}

// NewInternalServerErrorWithError wraps an existing error while creating a new empty client message and a InternalServerError (500) return code
func NewInternalServerErrorWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: internalServerError,
		error:   err,
	})
}

// NewServiceUnavailableWithError wraps an existing error while creating a new empty client message and a ServiceUnavailable (503) return code
func NewServiceUnavailableWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: serviceUnavailable,
		error:   err,
	})
}

// NewBadRequestMessage creates a new client message with a BadRequest (400) return code
func NewBadRequestMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       badRequest,
		clientMessage: message,
	})
}

// NewUnauthorizedMessage creates a new client message with a Unauthorized (401) return code
func NewUnauthorizedMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: message,
	})
}

// NewForbiddenMessage creates a new client message with a Forbidden (403) return code
func NewForbiddenMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       forbidden,
		clientMessage: message,
	})
}

// NewNotFoundMessage creates a new client message with a NotFound (404) return code
func NewNotFoundMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       notFound,
		clientMessage: message,
	})
}

// NewConflictMessage creates a new client message with a Conflict (409) return code
func NewConflictMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       conflict,
		clientMessage: message,
	})
}

// NewInternalServerErrorMessage creates a new client message with a InternalServerError (500) return code
func NewInternalServerErrorMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: message,
	})
}

// NewServiceUnavailableMessage creates a new client message with a ServiceUnavailable (503) return code
func NewServiceUnavailableMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: message,
	})
}

// NewBadRequestMessagef creates a new client message with a BadRequest (400) return code
func NewBadRequestMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       badRequest,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewUnauthorizedMessagef creates a new client message with a Unauthorized (401) return code
func NewUnauthorizedMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewForbiddenMessagef creates a new client message with a Forbidden (403) return code
func NewForbiddenMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       forbidden,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewNotFoundMessagef creates a new client message with a NotFound (404) return code
func NewNotFoundMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       notFound,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewConflictMessagef creates a new client message with a Conflict (409) return code
func NewConflictMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       conflict,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewInternalServerErrorMessagef creates a new client message with a InternalServerError (500) return code
func NewInternalServerErrorMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewServiceUnavailableMessagef creates a new client message with a ServiceUnavailable (503) return code
func NewServiceUnavailableMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewBadRequestMessageWithError wraps an existing error while creating a new client message with a BadRequest (400) return code
func NewBadRequestMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       badRequest,
		clientMessage: message,
		error:         err,
	})
}

// NewUnauthorizedMessageWithError wraps an existing error while creating a new client message with a Unauthorized (401) return code
func NewUnauthorizedMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: message,
		error:         err,
	})
}

// NewForbiddenMessageWithError wraps an existing error while creating a new client message with a Forbidden (403) return code
func NewForbiddenMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       forbidden,
		clientMessage: message,
		error:         err,
	})
}

// NewNotFoundMessageWithError wraps an existing error while creating a new client message with a NotFound (404) return code
func NewNotFoundMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       notFound,
		clientMessage: message,
		error:         err,
	})
}

// NewConflictMessageWithError wraps an existing error while creating a new client message with a Conflict (409) return code
func NewConflictMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       conflict,
		clientMessage: message,
		error:         err,
	})
}

// NewInternalServerErrorMessageWithError wraps an existing error while creating a new client message with a InternalServerError (500) return code
func NewInternalServerErrorMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: message,
		error:         err,
	})
}

// NewServiceUnavailableMessageWithError wraps an existing error while creating a new client message with a ServiceUnavailable (503) return code
func NewServiceUnavailableMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: message,
		error:         err,
	})
}

// NewBadRequestMessageWithErrorf wraps an existing error while creating a new client message with a BadRequest (400) return code
func NewBadRequestMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       badRequest,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewUnauthorizedMessageWithErrorf wraps an existing error while creating a new client message with a Unauthorized (401) return code
func NewUnauthorizedMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewForbiddenMessageWithErrorf wraps an existing error while creating a new client message with a Forbidden (403) return code
func NewForbiddenMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       forbidden,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewNotFoundMessageWithErrorf wraps an existing error while creating a new client message with a NotFound (404) return code
func NewNotFoundMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       notFound,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewConflictMessageWithErrorf wraps an existing error while creating a new client message with a Conflict (409) return code
func NewConflictMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       conflict,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewInternalServerErrorMessageWithErrorf wraps an existing error while creating a new client message with a InternalServerError (500) return code
func NewInternalServerErrorMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewServiceUnavailableMessageWithErrorf wraps an existing error while creating a new client message with a ServiceUnavailable (503) return code
func NewServiceUnavailableMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// HasClientMessage checks if the error contains a client message
func HasClientMessage(err error) bool {
	cerr := &ClientMessage{}

	return errors.As(err, &cerr)
}

// HasBadRequest checks if the error contains a BadRequest (400) message
func HasBadRequest(err error) bool {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		return cerr.msgType == badRequest
	}

	return false
}

// HasUnauthorized checks if the error contains an Unauthorized (401) message
func HasUnauthorized(err error) bool {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		return cerr.msgType == unauthorized
	}

	return false
}

// HasForbidden checks if the error contains a Forbidden (403) message
func HasForbidden(err error) bool {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		return cerr.msgType == forbidden
	}

	return false
}

// HasNotFound checks if the error contains a NotFound (404) message
func HasNotFound(err error) bool {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		return cerr.msgType == notFound
	}

	return false
}

// HasConflict checks if the error contains a Conflict (409) message
func HasConflict(err error) bool {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		return cerr.msgType == conflict
	}

	return false
}

// HasInternalServerError checks if the error contains an InternalServerError (500) message
func HasInternalServerError(err error) bool {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		return cerr.msgType == internalServerError
	}

	return false
}

// HasServiceUnavailable checks if the error contains a ServiceUnavailable (503) message
func HasServiceUnavailable(err error) bool {
	cerr := &ClientMessage{}
	if errors.As(err, &cerr) {
		return cerr.msgType == serviceUnavailable
	}

	return false
}
