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
			}
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
	return c.clientMessage
}

func (c *ClientMessage) Unwrap() error {
	return c.error
}

func wrap(err error) errors.Chain {
	return errors.WrapSkipFrames(err, "", 3)
}

// NewBadRequest creates a new empty client message with a BadRequest (400) return code
func NewBadRequest() errors.Chain {
	return newBadRequest()
}

func newBadRequest() errors.Chain {
	return wrap(&ClientMessage{
		msgType: badRequest,
	})
}

// NewUnauthorized creates a new empty client message with a Unauthorized (401) return code
func NewUnauthorized() errors.Chain {
	return newUnauthorized()
}

func newUnauthorized() errors.Chain {
	return wrap(&ClientMessage{
		msgType: unauthorized,
	})
}

// NewForbidden creates a new empty client message with a Forbidden (403) return code
func NewForbidden() errors.Chain {
	return newForbidden()
}

func newForbidden() errors.Chain {
	return wrap(&ClientMessage{
		msgType: forbidden,
	})
}

// NewNotFound creates a new empty client message with a NotFound (404) return code
func NewNotFound() errors.Chain {
	return newNotFound()
}

func newNotFound() errors.Chain {
	return wrap(&ClientMessage{
		msgType: notFound,
	})
}

// NewConflict creates a new empty client message with a Conflict (409) return code
func NewConflict() errors.Chain {
	return newConflict()
}

func newConflict() errors.Chain {
	return wrap(&ClientMessage{
		msgType: conflict,
	})
}

// NewInternalServerError creates a new empty client message with a InternalServerError (500) return code
func NewInternalServerError() errors.Chain {
	return newInternalServerError()
}

func newInternalServerError() errors.Chain {
	return wrap(&ClientMessage{
		msgType: internalServerError,
	})
}

// NewServiceUnavailable creates a new empty client message with a ServiceUnavailable (503) return code
func NewServiceUnavailable() errors.Chain {
	return newServiceUnavailable()
}

func newServiceUnavailable() errors.Chain {
	return wrap(&ClientMessage{
		msgType: serviceUnavailable,
	})
}

// NewBadRequestWithError creates a new empty client message with error and a BadRequest (400) return code
func NewBadRequestWithError(err error) errors.Chain {
	return newBadRequestWithError(err)
}

func newBadRequestWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: badRequest,
		error:   err,
	})
}

// NewUnauthorizedWithError creates a new empty client message with error and a Unauthorized (401) return code
func NewUnauthorizedWithError(err error) errors.Chain {
	return newUnauthorizedWithError(err)
}

func newUnauthorizedWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: unauthorized,
		error:   err,
	})
}

// NewForbiddenWithError creates a new empty client message with error and a Forbidden (403) return code
func NewForbiddenWithError(err error) errors.Chain {
	return newForbiddenWithError(err)
}

func newForbiddenWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: forbidden,
		error:   err,
	})
}

// NewNotFoundWithError creates a new empty client message with error and a NotFound (404) return code
func NewNotFoundWithError(err error) errors.Chain {
	return newNotFoundWithError(err)
}

func newNotFoundWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: notFound,
		error:   err,
	})
}

// NewConflictWithError creates a new empty client message with error and a Conflict (409) return code
func NewConflictWithError(err error) errors.Chain {
	return newConflictWithError(err)
}

func newConflictWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: conflict,
		error:   err,
	})
}

// NewInternalServerErrorWithError creates a new empty client message with error and a InternalServerError (500) return code
func NewInternalServerErrorWithError(err error) errors.Chain {
	return newInternalServerErrorWithError(err)
}

func newInternalServerErrorWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: internalServerError,
		error:   err,
	})
}

// NewServiceUnavailableWithError creates a new empty client message with error and a ServiceUnavailable (503) return code
func NewServiceUnavailableWithError(err error) errors.Chain {
	return newServiceUnavailableWithError(err)
}

func newServiceUnavailableWithError(err error) errors.Chain {
	return wrap(&ClientMessage{
		msgType: serviceUnavailable,
		error:   err,
	})
}

// NewBadRequestMessage creates a new client message with a BadRequest (400) return code
func NewBadRequestMessage(message string) errors.Chain {
	return newBadRequestMessage(message)
}

func newBadRequestMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       badRequest,
		clientMessage: message,
	})
}

// NewUnauthorizedMessage creates a new client message with a Unauthorized (401) return code
func NewUnauthorizedMessage(message string) errors.Chain {
	return newUnauthorizedMessage(message)
}

func newUnauthorizedMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: message,
	})
}

// NewForbiddenMessage creates a new client message with a Forbidden (403) return code
func NewForbiddenMessage(message string) errors.Chain {
	return newForbiddenMessage(message)
}

func newForbiddenMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       forbidden,
		clientMessage: message,
	})
}

// NewNotFoundMessage creates a new client message with a NotFound (404) return code
func NewNotFoundMessage(message string) errors.Chain {
	return newNotFoundMessage(message)
}

func newNotFoundMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       notFound,
		clientMessage: message,
	})
}

// NewConflictMessage creates a new client message with a Conflict (409) return code
func NewConflictMessage(message string) errors.Chain {
	return newConflictMessage(message)
}

func newConflictMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       conflict,
		clientMessage: message,
	})
}

// NewInternalServerErrorMessage creates a new client message with a InternalServerError (500) return code
func NewInternalServerErrorMessage(message string) errors.Chain {
	return newInternalServerErrorMessage(message)
}

func newInternalServerErrorMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: message,
	})
}

// NewServiceUnavailableMessage creates a new client message with a ServiceUnavailable (503) return code
func NewServiceUnavailableMessage(message string) errors.Chain {
	return newServiceUnavailableMessage(message)
}

func newServiceUnavailableMessage(message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: message,
	})
}

// NewBadRequestMessagef creates a new client message with a BadRequest (400) return code
func NewBadRequestMessagef(format string, a ...any) errors.Chain {
	return newBadRequestMessagef(format, a...)
}

func newBadRequestMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       badRequest,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewUnauthorizedMessagef creates a new client message with a Unauthorized (401) return code
func NewUnauthorizedMessagef(format string, a ...any) errors.Chain {
	return newUnauthorizedMessagef(format, a...)
}

func newUnauthorizedMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewForbiddenMessagef creates a new client message with a Forbidden (403) return code
func NewForbiddenMessagef(format string, a ...any) errors.Chain {
	return newForbiddenMessagef(format, a...)
}

func newForbiddenMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       forbidden,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewNotFoundMessagef creates a new client message with a NotFound (404) return code
func NewNotFoundMessagef(format string, a ...any) errors.Chain {
	return newNotFoundMessagef(format, a...)
}

func newNotFoundMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       notFound,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewConflictMessagef creates a new client message with a Conflict (409) return code
func NewConflictMessagef(format string, a ...any) errors.Chain {
	return newConflictMessagef(format, a...)
}

func newConflictMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       conflict,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewInternalServerErrorMessagef creates a new client message with a InternalServerError (500) return code
func NewInternalServerErrorMessagef(format string, a ...any) errors.Chain {
	return newInternalServerErrorMessagef(format, a...)
}

func newInternalServerErrorMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewServiceUnavailableMessagef creates a new client message with a ServiceUnavailable (503) return code
func NewServiceUnavailableMessagef(format string, a ...any) errors.Chain {
	return newServiceUnavailableMessagef(format, a...)
}

func newServiceUnavailableMessagef(format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: fmt.Sprintf(format, a...),
	})
}

// NewBadRequestMessageWithError wraps an existing error while creating a new client message with a BadRequest (400) return code
func NewBadRequestMessageWithError(err error, message string) errors.Chain {
	return newBadRequestMessageWithError(err, message)
}

func newBadRequestMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       badRequest,
		clientMessage: message,
		error:         err,
	})
}

// NewUnauthorizedMessageWithError wraps an existing error while creating a new client message with a Unauthorized (401) return code
func NewUnauthorizedMessageWithError(err error, message string) errors.Chain {
	return newUnauthorizedMessageWithError(err, message)
}

func newUnauthorizedMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: message,
		error:         err,
	})
}

// NewForbiddenMessageWithError wraps an existing error while creating a new client message with a Forbidden (403) return code
func NewForbiddenMessageWithError(err error, message string) errors.Chain {
	return newForbiddenMessageWithError(err, message)
}

func newForbiddenMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       forbidden,
		clientMessage: message,
		error:         err,
	})
}

// NewNotFoundMessageWithError wraps an existing error while creating a new client message with a NotFound (404) return code
func NewNotFoundMessageWithError(err error, message string) errors.Chain {
	return newNotFoundMessageWithError(err, message)
}

func newNotFoundMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       notFound,
		clientMessage: message,
		error:         err,
	})
}

// NewConflictMessageWithError wraps an existing error while creating a new client message with a Conflict (409) return code
func NewConflictMessageWithError(err error, message string) errors.Chain {
	return newConflictMessageWithError(err, message)
}

func newConflictMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       conflict,
		clientMessage: message,
		error:         err,
	})
}

// NewInternalServerErrorMessageWithError wraps an existing error while creating a new client message with a InternalServerError (500) return code
func NewInternalServerErrorMessageWithError(err error, message string) errors.Chain {
	return newInternalServerErrorMessageWithError(err, message)
}

func newInternalServerErrorMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: message,
		error:         err,
	})
}

// NewServiceUnavailableMessageWithError wraps an existing error while creating a new client message with a ServiceUnavailable (503) return code
func NewServiceUnavailableMessageWithError(err error, message string) errors.Chain {
	return newServiceUnavailableMessageWithError(err, message)
}

func newServiceUnavailableMessageWithError(err error, message string) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       serviceUnavailable,
		clientMessage: message,
		error:         err,
	})
}

// NewBadRequestMessageWithErrorf wraps an existing error while creating a new client message with a BadRequest (400) return code
func NewBadRequestMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return newBadRequestMessageWithErrorf(err, format, a...)
}

func newBadRequestMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       badRequest,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewUnauthorizedMessageWithErrorf wraps an existing error while creating a new client message with a Unauthorized (401) return code
func NewUnauthorizedMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return newUnauthorizedMessageWithErrorf(err, format, a...)
}

func newUnauthorizedMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       unauthorized,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewForbiddenMessageWithErrorf wraps an existing error while creating a new client message with a Forbidden (403) return code
func NewForbiddenMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return newForbiddenMessageWithErrorf(err, format, a...)
}

func newForbiddenMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       forbidden,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewNotFoundMessageWithErrorf wraps an existing error while creating a new client message with a NotFound (404) return code
func NewNotFoundMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return newNotFoundMessageWithErrorf(err, format, a...)
}

func newNotFoundMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       notFound,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewConflictMessageWithErrorf wraps an existing error while creating a new client message with a Conflict (409) return code
func NewConflictMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return newConflictMessageWithErrorf(err, format, a...)
}

func newConflictMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       conflict,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewInternalServerErrorMessageWithErrorf wraps an existing error while creating a new client message with a InternalServerError (500) return code
func NewInternalServerErrorMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return newInternalServerErrorMessageWithErrorf(err, format, a...)
}

func newInternalServerErrorMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return wrap(&ClientMessage{
		msgType:       internalServerError,
		clientMessage: fmt.Sprintf(format, a...),
		error:         err,
	})
}

// NewServiceUnavailableMessageWithErrorf wraps an existing error while creating a new client message with a ServiceUnavailable (503) return code
func NewServiceUnavailableMessageWithErrorf(err error, format string, a ...any) errors.Chain {
	return newServiceUnavailableMessageWithErrorf(err, format, a...)
}

func newServiceUnavailableMessageWithErrorf(err error, format string, a ...any) errors.Chain {
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
