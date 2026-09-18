package httpio

import (
	"encoding"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/cccteam/ccc"
	"github.com/cccteam/logger"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
)

// ParamType defines the type used to describe url Params
type ParamType string

// paramErrMsg is the panic value raised by Param when a route parameter is missing or
// fails to parse. msg is safe to return to the client. err carries the underlying parse
// failure, which may include internal source paths and library names, and must only
// ever be logged server-side.
type paramErrMsg struct {
	msg string
	err error
}

func newParamErrMsg(err error, format string, a ...any) paramErrMsg {
	return paramErrMsg{msg: fmt.Sprintf(format, a...), err: err}
}

// Msg returns the client-safe description of the parameter failure
func (m paramErrMsg) Msg() string {
	return m.msg
}

// Err returns the underlying parse error, or nil if the parameter was simply missing
func (m paramErrMsg) Err() error {
	return m.err
}

// WithParams middleware is used to capture Param Parsing errors. They are returned
// as a http.StatusBadRequest status code with a generic message naming the offending
// parameter. The underlying parse error is logged server-side and never sent to the client.
func WithParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				m, ok := rec.(paramErrMsg)
				if !ok {
					panic(rec)
				}

				if err := NewEncoder(w).BadRequestMessageWithError(r.Context(), m.Err(), m.Msg()); err != nil {
					logger.FromReq(r).Info(err)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Param extracts the Param from the Request Context
func Param[T any](r *http.Request, param ParamType) (val T) {
	fetchParam := func(r *http.Request, param ParamType) any {
		v := chi.URLParam(r, string(param))
		if v == "" {
			panic(newParamErrMsg(nil, "route parameter (%s) is required", param))
		}
		switch any(val).(type) {
		case string:
			return v
		case int:
			i, err := strconv.Atoi(v)
			if err != nil {
				panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
			}

			return i
		case int64:
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
			}

			return i
		case float64:
			i, err := strconv.ParseFloat(v, 64)
			if err != nil {
				panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
			}

			return i
		case bool:
			i, err := strconv.ParseBool(v)
			if err != nil {
				panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
			}

			return i
		case uuid.UUID:
			i, err := uuid.FromString(v)
			if err != nil {
				panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
			}

			return i
		case ccc.UUID:
			i, err := ccc.UUIDFromString(v)
			if err != nil {
				panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
			}

			return i
		default:
			if val2 := resolveInterfaces(param, v, val); val2 != nil {
				return val2
			}

			// handle named types
			rt := reflect.TypeOf(val)
			switch rt.Kind() {
			case reflect.String:
				return reflect.ValueOf(v).Convert(rt).Interface()
			case reflect.Int:
				i, err := strconv.Atoi(v)
				if err != nil {
					panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
				}

				return reflect.ValueOf(i).Convert(rt).Interface()
			case reflect.Int64:
				i, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
				}

				return reflect.ValueOf(i).Convert(rt).Interface()
			case reflect.Float64:
				i, err := strconv.ParseFloat(v, 64)
				if err != nil {
					panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
				}

				return reflect.ValueOf(i).Convert(rt).Interface()
			case reflect.Bool:
				i, err := strconv.ParseBool(v)
				if err != nil {
					panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
				}

				return reflect.ValueOf(i).Convert(rt).Interface()
			default:
				if rt.ConvertibleTo(reflect.TypeOf(uuid.UUID{})) {
					i, err := uuid.FromString(v)
					if err != nil {
						panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
					}

					return reflect.ValueOf(i).Convert(rt).Interface()
				}

				panic(fmt.Sprintf("support for %T has not been implemented", val))
			}
		}
	}

	v := fetchParam(r, param)
	val, ok := v.(T)
	if !ok {
		panic(fmt.Sprintf("implementation error: returned %T instead of %T", v, val))
	}

	return val
}

func resolveInterfaces[T any](param ParamType, paramVal string, val T) any {
	var receivedPtr bool
	var val2 any

	// We need a pointer because these interfaces are implemented on pointer receivers
	t := reflect.TypeOf(val)
	if t.Kind() == reflect.Pointer {
		receivedPtr = true
		// In this case, T is a nil pointer
		val2 = reflect.New(t.Elem()).Interface().(T)
	} else {
		val2 = &val
	}

	switch t := val2.(type) {
	case encoding.TextUnmarshaler:
		if err := t.UnmarshalText([]byte(paramVal)); err != nil {
			panic(newParamErrMsg(err, "route parameter (%s) is not a valid %T", param, val))
		}
	default:
		return nil
	}

	if receivedPtr {
		return val2
	}

	return *(val2.(*T))
}
