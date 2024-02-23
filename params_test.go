package httpio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
)

func TestWithParams(t *testing.T) {
	t.Parallel()

	type args struct {
		h http.Handler
	}
	tests := []struct {
		name      string
		args      args
		wantPanic bool
		wantCode  int
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
					panic(paramErrMsg("message"))
				}),
			},
			wantCode: http.StatusBadRequest,
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if tt.wantPanic != (r != nil) {
					t.Errorf("param() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			req := mockRequest(map[ParamType]string{})
			rr := httptest.NewRecorder()

			WithParams(tt.args.h).ServeHTTP(rr, req)

			if code := rr.Code; code != tt.wantCode {
				t.Errorf("WithParam() code = %v, want %v", code, tt.wantCode)
			}
		})
	}
}

func Test_param_string(t *testing.T) {
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
		tt := tt
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

func Test_param_int(t *testing.T) {
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
		tt := tt
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

func Test_param_int64(t *testing.T) {
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
		tt := tt
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

func Test_param_float64(t *testing.T) {
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
		tt := tt
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

func Test_param_bool(t *testing.T) {
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
		tt := tt
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

func Test_param_UUID(t *testing.T) {
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
		tt := tt
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

func Test_param_ptr_UUID(t *testing.T) {
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
		tt := tt
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

func Test_param_notimplemented(t *testing.T) {
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
		tt := tt
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
