package httpio

import (
	stderr "errors"
	"testing"

	"github.com/go-playground/errors/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestMessage(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "find messages",
			args: args{&ClientMessage{clientMessage: "my message"}},
			want: "my message",
		},
		{
			name: "dont find messages",
			args: args{errors.New("my message")},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Message(tt.args.err); got != tt.want {
				t.Errorf("Message() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessages(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "find message",
			args: args{&ClientMessage{clientMessage: "my message"}},
			want: []string{"my message"},
		},
		{
			name: "find messages",
			args: args{&ClientMessage{clientMessage: "my message", error: errors.Wrap(&ClientMessage{clientMessage: "your message"}, "")}},
			want: []string{"my message", "your message"},
		},
		{
			name: "dont find messages",
			args: args{errors.New("my message")},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Messages(tt.args.err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Messages() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestContainsError(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ClientMessage error exists",
			args: args{&ClientMessage{error: stderr.New("my error")}},
			want: true,
		},
		{
			name: "ClientMessage error does not exist",
			args: args{err: &ClientMessage{clientMessage: "my message"}},
			want: false,
		},
		{
			name: "standard error",
			args: args{err: stderr.New("my error")},
			want: true,
		},
		{
			name: "nested ClientMessage with error",
			args: args{err: NewBadRequestMessageWithError(NewBadRequestMessageWithError(stderr.New("my error"), "first message"), "second message")},
			want: true,
		},
		{
			name: "nested ClientMessage with no error",
			args: args{err: NewBadRequestMessageWithError(NewBadRequestMessage("first message"), "second message")},
			want: false,
		},
		{
			name: "nil error",
			args: args{err: nil},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CauseIsError(tt.args.err); got != tt.want {
				t.Errorf("ContainsError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientMessage_Error(t *testing.T) {
	t.Parallel()

	type fields struct {
		msgType       msgType
		clientMessage string
		error         error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "with error",
			fields: fields{
				clientMessage: "my message",
				error:         stderr.New("my error"),
			},
			want: "my message",
		},
		{
			name: "no error",
			fields: fields{
				clientMessage: "my message",
			},
			want: "my message",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &ClientMessage{
				msgType:       tt.fields.msgType,
				clientMessage: tt.fields.clientMessage,
				error:         tt.fields.error,
			}
			if got := c.Error(); got != tt.want {
				t.Errorf("ClientMessage.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientMessage_Unwrap(t *testing.T) {
	t.Parallel()

	type fields struct {
		msgType       msgType
		clientMessage string
		error         error
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			fields: fields{
				error: stderr.New("my error"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &ClientMessage{
				msgType:       tt.fields.msgType,
				clientMessage: tt.fields.clientMessage,
				error:         tt.fields.error,
			}
			err := c.Unwrap()
			if diff := cmp.Diff(tt.fields.error, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("ClientMessage.Unwrap() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClientMessage_Message(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "BadRequest (with message)", args: args{err: NewBadRequestMessage("msg")}, want: "msg"},
		{name: "BadRequest (with messagef)", args: args{err: NewBadRequestMessagef("msg %v", "arg")}, want: "msg arg"},
		{name: "BadRequest (with message and error)", args: args{err: NewBadRequestMessageWithError(stderr.New("err"), "msg")}, want: "msg"},
		{name: "BadRequest (with message and errorf)", args: args{err: NewBadRequestMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: "msg arg"},
		{name: "Unauthorized (with message)", args: args{err: NewUnauthorizedMessage("msg")}, want: "msg"},
		{name: "Unauthorized (with messagef)", args: args{err: NewUnauthorizedMessagef("msg %v", "arg")}, want: "msg arg"},
		{name: "Unauthorized (with message and error)", args: args{err: NewUnauthorizedMessageWithError(stderr.New("err"), "msg")}, want: "msg"},
		{name: "Unauthorized (with message and errorf)", args: args{err: NewUnauthorizedMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: "msg arg"},
		{name: "Forbidden (with message)", args: args{err: NewForbiddenMessage("msg")}, want: "msg"},
		{name: "Forbidden (with messagef)", args: args{err: NewForbiddenMessagef("msg %v", "arg")}, want: "msg arg"},
		{name: "Forbidden (with message and error)", args: args{err: NewForbiddenMessageWithError(stderr.New("err"), "msg")}, want: "msg"},
		{name: "Forbidden (with message and errorf)", args: args{err: NewForbiddenMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: "msg arg"},
		{name: "NotFound (with message)", args: args{err: NewNotFoundMessage("msg")}, want: "msg"},
		{name: "NotFound (with messagef)", args: args{err: NewNotFoundMessagef("msg %v", "arg")}, want: "msg arg"},
		{name: "NotFound (with message and error)", args: args{err: NewNotFoundMessageWithError(stderr.New("err"), "msg")}, want: "msg"},
		{name: "NotFound (with message and errorf)", args: args{err: NewNotFoundMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: "msg arg"},
		{name: "Conflict (with message)", args: args{err: NewConflictMessage("msg")}, want: "msg"},
		{name: "Conflict (with messagef)", args: args{err: NewConflictMessagef("msg %v", "arg")}, want: "msg arg"},
		{name: "Conflict (with message and error)", args: args{err: NewConflictMessageWithError(stderr.New("err"), "msg")}, want: "msg"},
		{name: "Conflict (with message and errorf)", args: args{err: NewConflictMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: "msg arg"},
		{name: "InternalServerError (with message)", args: args{err: NewInternalServerErrorMessage("msg")}, want: "msg"},
		{name: "InternalServerError (with messagef)", args: args{err: NewInternalServerErrorMessagef("msg %v", "arg")}, want: "msg arg"},
		{name: "InternalServerError (with message and error)", args: args{err: NewInternalServerErrorMessageWithError(stderr.New("err"), "msg")}, want: "msg"},
		{name: "InternalServerError (with message and errorf)", args: args{err: NewInternalServerErrorMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: "msg arg"},
		{name: "ServiceUnavailable (with message)", args: args{err: NewServiceUnavailableMessage("msg")}, want: "msg"},
		{name: "ServiceUnavailable (with messagef)", args: args{err: NewServiceUnavailableMessagef("msg %v", "arg")}, want: "msg arg"},
		{name: "ServiceUnavailable (with message and error)", args: args{err: NewServiceUnavailableMessageWithError(stderr.New("err"), "msg")}, want: "msg"},
		{name: "ServiceUnavailable (with message and errorf)", args: args{err: NewServiceUnavailableMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: "msg arg"},
		{name: "Other error", args: args{err: stderr.New("err")}, want: ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cerr := &ClientMessage{}
			stderr.As(tt.args.err, &cerr)
			if got := cerr.Message(); got != tt.want {
				t.Errorf("ClientMessage.Message() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasClientMessage(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "BadRequest", args: args{err: NewBadRequest()}, want: true},
		{name: "BadRequest (with error)", args: args{err: NewBadRequestWithError(stderr.New("msg"))}, want: true},
		{name: "BadRequest (with message)", args: args{err: NewBadRequestMessage("msg")}, want: true},
		{name: "BadRequest (with messagef)", args: args{err: NewBadRequestMessagef("msg %v", "arg")}, want: true},
		{name: "BadRequest (with message and error)", args: args{err: NewBadRequestMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "BadRequest (with message and errorf)", args: args{err: NewBadRequestMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},

		{name: "Unauthorized", args: args{err: NewUnauthorized()}, want: true},
		{name: "Unauthorized (with error)", args: args{err: NewUnauthorizedWithError(stderr.New("msg"))}, want: true},
		{name: "Unauthorized (with message)", args: args{err: NewUnauthorizedMessage("msg")}, want: true},
		{name: "Unauthorized (with messagef)", args: args{err: NewUnauthorizedMessagef("msg %v", "arg")}, want: true},
		{name: "Unauthorized (with message and error)", args: args{err: NewUnauthorizedMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "Unauthorized (with message and errorf)", args: args{err: NewUnauthorizedMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},

		{name: "Forbidden", args: args{err: NewForbidden()}, want: true},
		{name: "Forbidden (with error)", args: args{err: NewForbiddenWithError(stderr.New("msg"))}, want: true},
		{name: "Forbidden (with message)", args: args{err: NewForbiddenMessage("msg")}, want: true},
		{name: "Forbidden (with messagef)", args: args{err: NewForbiddenMessagef("msg %v", "arg")}, want: true},
		{name: "Forbidden (with message and error)", args: args{err: NewForbiddenMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "Forbidden (with message and errorf)", args: args{err: NewForbiddenMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},

		{name: "NotFound", args: args{err: NewNotFound()}, want: true},
		{name: "NotFound (with error)", args: args{err: NewNotFoundWithError(stderr.New("msg"))}, want: true},
		{name: "NotFound (with message)", args: args{err: NewNotFoundMessage("msg")}, want: true},
		{name: "NotFound (with messagef)", args: args{err: NewNotFoundMessagef("msg %v", "arg")}, want: true},
		{name: "NotFound (with message and error)", args: args{err: NewNotFoundMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "NotFound (with message and errorf)", args: args{err: NewNotFoundMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},

		{name: "Conflict", args: args{err: NewConflict()}, want: true},
		{name: "Conflict (with error)", args: args{err: NewConflictWithError(stderr.New("msg"))}, want: true},
		{name: "Conflict (with message)", args: args{err: NewConflictMessage("msg")}, want: true},
		{name: "Conflict (with messagef)", args: args{err: NewConflictMessagef("msg %v", "arg")}, want: true},
		{name: "Conflict (with message and error)", args: args{err: NewConflictMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "Conflict (with message and errorf)", args: args{err: NewConflictMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},

		{name: "InternalServerError", args: args{err: NewInternalServerError()}, want: true},
		{name: "InternalServerError (with error)", args: args{err: NewInternalServerErrorWithError(stderr.New("msg"))}, want: true},
		{name: "InternalServerError (with message)", args: args{err: NewInternalServerErrorMessage("msg")}, want: true},
		{name: "InternalServerError (with messagef)", args: args{err: NewInternalServerErrorMessagef("msg %v", "arg")}, want: true},
		{name: "InternalServerError (with message and error)", args: args{err: NewInternalServerErrorMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "InternalServerError (with message and errorf)", args: args{err: NewInternalServerErrorMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},

		{name: "ServiceUnavailable", args: args{err: NewServiceUnavailable()}, want: true},
		{name: "ServiceUnavailable (with error)", args: args{err: NewServiceUnavailableWithError(stderr.New("msg"))}, want: true},
		{name: "ServiceUnavailable (with message)", args: args{err: NewServiceUnavailableMessage("msg")}, want: true},
		{name: "ServiceUnavailable (with messagef)", args: args{err: NewServiceUnavailableMessagef("msg %v", "arg")}, want: true},
		{name: "ServiceUnavailable (with message and error)", args: args{err: NewServiceUnavailableMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "ServiceUnavailable (with message and errorf)", args: args{err: NewServiceUnavailableMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},

		{name: "Other error", args: args{err: stderr.New("err")}, want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasClientMessage(tt.args.err); got != tt.want {
				t.Errorf("HasClientMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasBadRequest(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "BadRequest (with message)", args: args{err: NewBadRequestMessage("msg")}, want: true},
		{name: "BadRequest (with messagef)", args: args{err: NewBadRequestMessagef("msg %v", "arg")}, want: true},
		{name: "BadRequest (with message and error)", args: args{err: NewBadRequestMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "BadRequest (with message and errorf)", args: args{err: NewBadRequestMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},
		{name: "Other error", args: args{err: stderr.New("err")}, want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasBadRequest(tt.args.err); got != tt.want {
				t.Errorf("HasBadRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasUnauthorized(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Unauthorized (with message)", args: args{err: NewUnauthorizedMessage("msg")}, want: true},
		{name: "Unauthorized (with messagef)", args: args{err: NewUnauthorizedMessagef("msg %v", "arg")}, want: true},
		{name: "Unauthorized (with message and error)", args: args{err: NewUnauthorizedMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "Unauthorized (with message and errorf)", args: args{err: NewUnauthorizedMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},
		{name: "Other error", args: args{err: stderr.New("err")}, want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasUnauthorized(tt.args.err); got != tt.want {
				t.Errorf("HasUnauthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasForbidden(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Forbidden (with message)", args: args{err: NewForbiddenMessage("msg")}, want: true},
		{name: "Forbidden (with messagef)", args: args{err: NewForbiddenMessagef("msg %v", "arg")}, want: true},
		{name: "Forbidden (with message and error)", args: args{err: NewForbiddenMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "Forbidden (with message and errorf)", args: args{err: NewForbiddenMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},
		{name: "Other error", args: args{err: stderr.New("err")}, want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasForbidden(tt.args.err); got != tt.want {
				t.Errorf("HasForbidden() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasNotFound(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "NotFound (with message)", args: args{err: NewNotFoundMessage("msg")}, want: true},
		{name: "NotFound (with messagef)", args: args{err: NewNotFoundMessagef("msg %v", "arg")}, want: true},
		{name: "NotFound (with message and error)", args: args{err: NewNotFoundMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "NotFound (with message and errorf)", args: args{err: NewNotFoundMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},
		{name: "Other error", args: args{err: stderr.New("err")}, want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasNotFound(tt.args.err); got != tt.want {
				t.Errorf("HasNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasConflict(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Conflict (with message)", args: args{err: NewConflictMessage("msg")}, want: true},
		{name: "Conflict (with messagef)", args: args{err: NewConflictMessagef("msg %v", "arg")}, want: true},
		{name: "Conflict (with message and error)", args: args{err: NewConflictMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "Conflict (with message and errorf)", args: args{err: NewConflictMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},
		{name: "Other error", args: args{err: stderr.New("err")}, want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasConflict(tt.args.err); got != tt.want {
				t.Errorf("HasConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasInternalServerError(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "InternalServerError (with message)", args: args{err: NewInternalServerErrorMessage("msg")}, want: true},
		{name: "InternalServerError (with messagef)", args: args{err: NewInternalServerErrorMessagef("msg %v", "arg")}, want: true},
		{name: "InternalServerError (with message and error)", args: args{err: NewInternalServerErrorMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "InternalServerError (with message and errorf)", args: args{err: NewInternalServerErrorMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},
		{name: "Other error", args: args{err: stderr.New("err")}, want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasInternalServerError(tt.args.err); got != tt.want {
				t.Errorf("HasInternalServerError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasServiceUnavailable(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "ServiceUnavailable (with message)", args: args{err: NewServiceUnavailableMessage("msg")}, want: true},
		{name: "ServiceUnavailable (with messagef)", args: args{err: NewServiceUnavailableMessagef("msg %v", "arg")}, want: true},
		{name: "ServiceUnavailable (with message and error)", args: args{err: NewServiceUnavailableMessageWithError(stderr.New("err"), "msg")}, want: true},
		{name: "ServiceUnavailable (with message and errorf)", args: args{err: NewServiceUnavailableMessageWithErrorf(stderr.New("err"), "msg %v", "arg")}, want: true},
		{name: "Other error", args: args{err: stderr.New("err")}, want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasServiceUnavailable(tt.args.err); got != tt.want {
				t.Errorf("HasServiceUnavailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
