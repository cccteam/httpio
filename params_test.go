package httpio

import (
	"context"
	"net/http"
	"net/http/httptest"
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
			wantVal: ccc.Ptr(ccc.Must(ccc.UUIDFromString("0020198f-a14e-42ee-b5f8-65a228ba38e7"))),
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
