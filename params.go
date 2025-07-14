package httpio

import (
	"encoding"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/cccteam/ccc"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
)

// ParamType defines the type used to describe url Params
type ParamType string

type paramErrMsg string

func newParamErrMsg(format string, a ...any) paramErrMsg {
	return paramErrMsg(fmt.Sprintf(format, a...))
}

func (m paramErrMsg) Msg() string {
	return string(m)
}

// WithParams middleware is used to capture Param Parsing errors. They are returned
// as a http.StatusBadRequest status code with a message describing any parsing issue
func WithParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if m, ok := rec.(paramErrMsg); ok {
					_ = NewEncoder(w).BadRequestMessage(r.Context(), m.Msg())
				} else {
					panic(rec)
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
			panic(newParamErrMsg("route parameter (%s) is required", param))
		}
		switch any(val).(type) {
		case string:
			return v
		case int:
			i, err := strconv.Atoi(v)
			if err != nil {
				panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
			}

			return i
		case int64:
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
			}

			return i
		case float64:
			i, err := strconv.ParseFloat(v, 64)
			if err != nil {
				panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
			}

			return i
		case bool:
			i, err := strconv.ParseBool(v)
			if err != nil {
				panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
			}

			return i
		case uuid.UUID:
			i, err := uuid.FromString(v)
			if err != nil {
				panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
			}

			return i
		case ccc.UUID:
			i, err := ccc.UUIDFromString(v)
			if err != nil {
				panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
			}

			return i
		default:
			if val2 := resolveInterfaces(param, v, val); val2 != nil {
				return val2
			}

			// handle named types
			switch reflect.TypeOf(val).Kind() {
			case reflect.String:
				return reflect.ValueOf(v).Convert(reflect.TypeOf(val)).Interface()
			case reflect.Int:
				i, err := strconv.Atoi(v)
				if err != nil {
					panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
				}

				return reflect.ValueOf(i).Convert(reflect.TypeOf(val)).Interface()
			case reflect.Int64:
				i, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
				}

				return reflect.ValueOf(i).Convert(reflect.TypeOf(val)).Interface()
			case reflect.Float64:
				i, err := strconv.ParseFloat(v, 64)
				if err != nil {
					panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
				}

				return reflect.ValueOf(i).Convert(reflect.TypeOf(val)).Interface()
			case reflect.Bool:
				i, err := strconv.ParseBool(v)
				if err != nil {
					panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
				}

				return reflect.ValueOf(i).Convert(reflect.TypeOf(val)).Interface()
			default:
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
			panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, paramVal, val, err))
		}
	default:
		return nil
	}

	if receivedPtr {
		return val2
	}

	return *(val2.(*T))
}
