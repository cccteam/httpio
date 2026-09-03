package httpio

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cccteam/ccc"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
)

func TestWithParams(t *testing.T) {
	t.Parallel()

	type args struct {
		h http.Handler
	}
	tests := []struct {
		name             string
		args             args
		wantPanic        bool
		wantCode         int
		wantBodyContains []string
		wantBodyExcludes []string
	}{
		{
			name: "success",
			args: args{
				h: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}),
			},
			wantCode: http.StatusOK,
		},
		{
			name: "No panic",
			args: args{
				h: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
					panic(newParamErrMsg(nil, "message"))
				}),
			},
			wantCode:         http.StatusBadRequest,
			wantBodyContains: []string{"message"},
		},
		{
			name: "parse failure hides underlying error",
			args: args{
				h: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
					panic(newParamErrMsg(errors.New("source=github.com/example/lib/parse.go:35: bad input"), "route parameter (id) is not a valid int"))
				}),
			},
			wantCode:         http.StatusBadRequest,
			wantBodyContains: []string{"route parameter (id) is not a valid int"},
			wantBodyExcludes: []string{"source=", "parse.go", "bad input"},
		},
		{
			name: "malformed ccc.UUID route parameter",
			args: args{
				h: http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
					_ = Param[ccc.UUID](r, "recordID")
				}),
			},
			wantCode:         http.StatusBadRequest,
			wantBodyContains: []string{"route parameter (recordID) is not a valid ccc.UUID"},
			wantBodyExcludes: []string{"source=", "uuid.go", "not-a-uuid"},
		},
		{
			name: "panic",
			args: args{
				h: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
					panic("message")
				}),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			req := mockRequest(map[ParamType]string{"recordID": "not-a-uuid"})
			rr := httptest.NewRecorder()

			WithParams(tt.args.h).ServeHTTP(rr, req)

			if code := rr.Code; code != tt.wantCode {
				t.Errorf("WithParam() code = %v, want %v", code, tt.wantCode)
			}
			body := rr.Body.String()
			for _, want := range tt.wantBodyContains {
				if !strings.Contains(body, want) {
					t.Errorf("WithParam() body = %q, want it to contain %q", body, want)
				}
			}
			for _, leaked := range tt.wantBodyExcludes {
				if strings.Contains(body, leaked) {
					t.Errorf("WithParam() body = %q, must not contain %q", body, leaked)
				}
			}
		})
	}
}

func TestParam_errMsg(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
		call  func(r *http.Request, param ParamType)
	}
	tests := []struct {
		name        string
		args        args
		wantMsg     string
		wantErr     bool
		wantExclude []string
	}{
		{
			name: "missing parameter has no underlying error",
			args: args{
				r:     mockRequest(map[ParamType]string{}),
				param: "id",
				call:  func(r *http.Request, param ParamType) { _ = Param[int](r, param) },
			},
			wantMsg: "route parameter (id) is required",
		},
		{
			name: "invalid int keeps parse error out of message",
			args: args{
				r:     mockRequest(map[ParamType]string{"id": "abc"}),
				param: "id",
				call:  func(r *http.Request, param ParamType) { _ = Param[int](r, param) },
			},
			wantMsg:     "route parameter (id) is not a valid int",
			wantErr:     true,
			wantExclude: []string{"abc", "strconv"},
		},
		{
			name: "invalid ccc.UUID keeps source path out of message",
			args: args{
				r:     mockRequest(map[ParamType]string{"id": "not-a-uuid"}),
				param: "id",
				call:  func(r *http.Request, param ParamType) { _ = Param[ccc.UUID](r, param) },
			},
			wantMsg:     "route parameter (id) is not a valid ccc.UUID",
			wantErr:     true,
			wantExclude: []string{"not-a-uuid", "source=", "uuid.go"},
		},
		{
			name: "invalid TextUnmarshaler keeps parse error out of message",
			args: args{
				r:     mockRequest(map[ParamType]string{"id": "not-a-uuid"}),
				param: "id",
				call:  func(r *http.Request, param ParamType) { _ = Param[*uuid.UUID](r, param) },
			},
			wantMsg:     "route parameter (id) is not a valid *uuid.UUID",
			wantErr:     true,
			wantExclude: []string{"not-a-uuid"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got paramErrMsg
			func() {
				defer func() {
					rec := recover()
					m, ok := rec.(paramErrMsg)
					if !ok {
						t.Fatalf("Param() panic = %v, want paramErrMsg", rec)
					}
					got = m
				}()
				tt.args.call(tt.args.r, tt.args.param)
			}()

			if got.Msg() != tt.wantMsg {
				t.Errorf("Param() Msg() = %q, want %q", got.Msg(), tt.wantMsg)
			}
			if (got.Err() != nil) != tt.wantErr {
				t.Errorf("Param() Err() = %v, wantErr %v", got.Err(), tt.wantErr)
			}
			for _, leaked := range tt.wantExclude {
				if strings.Contains(got.Msg(), leaked) {
					t.Errorf("Param() Msg() = %q, must not contain %q", got.Msg(), leaked)
				}
			}
		})
	}
}

func TestParam_named_string(t *testing.T) {
	t.Parallel()

	type NamedType string

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   NamedType
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "755"}),
				param: ParamType("guarantorId"),
			},
			wantVal: "755",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[NamedType](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_named_int(t *testing.T) {
	t.Parallel()

	type Namedtype int

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   Namedtype
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "755"}),
				param: ParamType("guarantorId"),
			},
			wantVal: 755,
		},
		{
			name: "Empty Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "-"}),
				param: ParamType("guarantorId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[Namedtype](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_named_int64(t *testing.T) {
	t.Parallel()

	type Namedtype int64

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   Namedtype
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "755"}),
				param: ParamType("guarantorId"),
			},
			wantVal: 755,
		},
		{
			name: "Empty Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "-"}),
				param: ParamType("guarantorId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[Namedtype](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_named_float64(t *testing.T) {
	t.Parallel()

	type Namedtype float64

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   Namedtype
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "755.01"}),
				param: ParamType("guarantorId"),
			},
			wantVal: 755.01,
		},
		{
			name: "Empty Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "-"}),
				param: ParamType("guarantorId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[Namedtype](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_named_ccc_uuid(t *testing.T) {
	t.Parallel()

	type NamedType ccc.UUID

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   NamedType
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "0020198f-a14e-42ee-b5f8-65a228ba38e7"}),
				param: ParamType("guarantorId"),
			},
			wantVal: NamedType(ccc.Must(ccc.UUIDFromString("0020198f-a14e-42ee-b5f8-65a228ba38e7"))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[NamedType](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_named_uuid_uuid(t *testing.T) {
	t.Parallel()

	type NamedType uuid.UUID

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   NamedType
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "0020198f-a14e-42ee-b5f8-65a228ba38e7"}),
				param: ParamType("guarantorId"),
			},
			wantVal: NamedType(ccc.Must(uuid.FromString("0020198f-a14e-42ee-b5f8-65a228ba38e7"))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[NamedType](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_named_bool(t *testing.T) {
	t.Parallel()

	type Namedtype bool

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   Namedtype
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "t"}),
				param: ParamType("guarantorId"),
			},
			wantVal: true,
		},
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "T"}),
				param: ParamType("guarantorId"),
			},
			wantVal: true,
		},
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "True"}),
				param: ParamType("guarantorId"),
			},
			wantVal: true,
		},
		{
			name: "Empty Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "-"}),
				param: ParamType("guarantorId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[Namedtype](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_string(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   string
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"guarantorId": "755"}),
				param: ParamType("guarantorId"),
			},
			wantVal: "755",
		},
		{
			name: "Empty Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{}),
				param: ParamType("guarantorId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[string](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_int(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   int
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "12"}),
				param: ParamType("fileId"),
			},
			wantVal: 12,
		},
		{
			name: "Invalid Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "12x"}),
				param: ParamType("fileId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[int](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_int64(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   int64
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "12"}),
				param: ParamType("fileId"),
			},
			wantVal: 12,
		},
		{
			name: "Invalid Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "12x"}),
				param: ParamType("fileId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[int64](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_float64(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   float64
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "12.34"}),
				param: ParamType("fileId"),
			},
			wantVal: 12.34,
		},
		{
			name: "Invalid Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "12.0x"}),
				param: ParamType("fileId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[float64](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_bool(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   bool
		wantPanic bool
	}{
		{
			name: "true",
			args: args{
				r:     mockRequest(map[ParamType]string{"active": "true"}),
				param: ParamType("active"),
			},
			wantVal: true,
		},
		{
			name: "t",
			args: args{
				r:     mockRequest(map[ParamType]string{"active": "t"}),
				param: ParamType("active"),
			},
			wantVal: true,
		},
		{
			name: "false",
			args: args{
				r:     mockRequest(map[ParamType]string{"active": "false"}),
				param: ParamType("active"),
			},
			wantVal: false,
		},
		{
			name: "f",
			args: args{
				r:     mockRequest(map[ParamType]string{"active": "f"}),
				param: ParamType("active"),
			},
			wantVal: false,
		},
		{
			name: "1",
			args: args{
				r:     mockRequest(map[ParamType]string{"active": "1"}),
				param: ParamType("active"),
			},
			wantVal: true,
		},
		{
			name: "0",
			args: args{
				r:     mockRequest(map[ParamType]string{"active": "0"}),
				param: ParamType("active"),
			},
			wantVal: false,
		},
		{
			name: "Invalid",
			args: args{
				r:     mockRequest(map[ParamType]string{"active": "x"}),
				param: ParamType("active"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[bool](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_UUID(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   uuid.UUID
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "0020198f-a14e-42ee-b5f8-65a228ba38e7"}),
				param: ParamType("fileId"),
			},
			wantVal: uuid.FromStringOrNil("0020198f-a14e-42ee-b5f8-65a228ba38e7"),
		},
		{
			name: "Invalid Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "0020198f-a14e-42ee-b5f8-65a228ba38xx"}),
				param: ParamType("fileId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[uuid.UUID](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_ptr_UUID(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   uuid.UUID
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "0020198f-a14e-42ee-b5f8-65a228ba38e7"}),
				param: ParamType("fileId"),
			},
			wantVal: uuid.FromStringOrNil("0020198f-a14e-42ee-b5f8-65a228ba38e7"),
		},
		{
			name: "Invalid Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "0020198f-a14e-42ee-b5f8-65a228ba38xx"}),
				param: ParamType("fileId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[*uuid.UUID](tt.args.r, tt.args.param); *gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_ccc_UUID(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   ccc.UUID
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "0020198f-a14e-42ee-b5f8-65a228ba38e7"}),
				param: ParamType("fileId"),
			},
			wantVal: ccc.Must(ccc.UUIDFromString("0020198f-a14e-42ee-b5f8-65a228ba38e7")),
		},
		{
			name: "Invalid Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "0020198f-a14e-42ee-b5f8-65a228ba38xx"}),
				param: ParamType("fileId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[ccc.UUID](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestParam_ptr_ccc_UUID(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   *ccc.UUID
		wantPanic bool
	}{
		{
			name: "Valid Param",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "0020198f-a14e-42ee-b5f8-65a228ba38e7"}),
				param: ParamType("fileId"),
			},
			wantVal: new(ccc.Must(ccc.UUIDFromString("0020198f-a14e-42ee-b5f8-65a228ba38e7"))),
		},
		{
			name: "Invalid Param Panic",
			args: args{
				r:     mockRequest(map[ParamType]string{"fileId": "0020198f-a14e-42ee-b5f8-65a228ba38xx"}),
				param: ParamType("fileId"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[*ccc.UUID](tt.args.r, tt.args.param); *gotVal != *tt.wantVal {
				t.Errorf("param() = %v, want %v", *gotVal, *tt.wantVal)
			}
		})
	}
}

func TestParam_notimplemented(t *testing.T) {
	t.Parallel()

	type args struct {
		r     *http.Request
		param ParamType
	}
	tests := []struct {
		name      string
		args      args
		wantVal   struct{}
		wantPanic bool
	}{
		{
			name: "Not implemented",
			args: args{
				r:     mockRequest(map[ParamType]string{"time": "2006-01-02"}),
				param: ParamType("time"),
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if gotVal := Param[struct{}](tt.args.r, tt.args.param); gotVal != tt.wantVal {
				t.Errorf("param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func Benchmark_param_int(b *testing.B) {
	r := mockRequest(map[ParamType]string{"integer": "1245"})

	b.ResetTimer()
	for range b.N {
		_ = Param[int](r, ParamType("integer"))
	}
}

func Benchmark_param_string(b *testing.B) {
	r := mockRequest(map[ParamType]string{"string": "755"})

	b.ResetTimer()
	for range b.N {
		_ = Param[string](r, ParamType("string"))
	}
}

func Benchmark_param_ccc_uuid(b *testing.B) {
	r := mockRequest(map[ParamType]string{"uuid": "0020198f-a14e-42ee-b5f8-65a228ba3899"})

	b.ResetTimer()
	for range b.N {
		_ = Param[ccc.UUID](r, ParamType("uuid"))
	}
}

func Benchmark_param_ccc_uuid_ptr(b *testing.B) {
	r := mockRequest(map[ParamType]string{"uuid": "0020198f-a14e-42ee-b5f8-65a228ba3899"})

	b.ResetTimer()
	for range b.N {
		_ = Param[*ccc.UUID](r, ParamType("uuid"))
	}
}

func Benchmark_param_uuid(b *testing.B) {
	r := mockRequest(map[ParamType]string{"uuid": "0020198f-a14e-42ee-b5f8-65a228ba3899"})

	b.ResetTimer()
	for range b.N {
		_ = Param[uuid.UUID](r, ParamType("uuid"))
	}
}

func Benchmark_param_uuid_ptr(b *testing.B) {
	r := mockRequest(map[ParamType]string{"uuid": "0020198f-a14e-42ee-b5f8-65a228ba3899"})

	b.ResetTimer()
	for range b.N {
		_ = Param[*uuid.UUID](r, ParamType("uuid"))
	}
}

func mockRequest(urlParams map[ParamType]string) *http.Request {
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "", http.NoBody)
	rctx := chi.NewRouteContext()
	for key, val := range urlParams {
		rctx.URLParams.Add(string(key), val)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return req
}
