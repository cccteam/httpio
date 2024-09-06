package httpio

import (
	"encoding"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-chi/chi/v5"
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
	fetchParam := func() any {
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
		default:
			t := reflect.TypeOf(val)
			if t.Kind() == reflect.Pointer {
				val = reflect.New(t.Elem()).Interface().(T)
				if resolveInterface(param, v, val) {
					return val
				}
			} else if resolveInterface(param, v, &val) {
				return val
			}

			panic(fmt.Sprintf("support for %T has not been implemented", val))
		}
	}

	v := fetchParam()
	val, ok := v.(T)
	if !ok {
		panic(fmt.Sprintf("implementation error: returned %T instead of %T", v, val))
	}

	return val
}

func resolveInterface(param ParamType, v string, val any) bool {
	switch t := val.(type) {
	case encoding.TextUnmarshaler:
		if err := t.UnmarshalText([]byte(v)); err != nil {
			panic(newParamErrMsg("param %s=%s is not a valid %T. err: %s", param, v, val, err))
		}

		return true
	}

	return false
}
