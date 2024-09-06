package httpio

import (
	"encoding"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/errors/v5"
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
		// Fetch the URL parameter
		urlParam := chi.URLParam(r, string(param))
		if urlParam == "" {
			panic(newParamErrMsg("route parameter (%s) is required", param))
		}

		switch t := any(val).(type) {
		case string:
			return urlParam
		case int:
			i, err := strconv.Atoi(urlParam)
			if err != nil {
				panic(errors.Wrapf(err, "urlParam %s is not a valid %T", urlParam, t))
			}

			return i
		case int64:
			i, err := strconv.ParseInt(urlParam, 10, 64)
			if err != nil {
				panic(errors.Wrapf(err, "urlParam %s is not a valid %T", urlParam, t))
			}

			return i
		case float64:
			i, err := strconv.ParseFloat(urlParam, 64)
			if err != nil {
				panic(errors.Wrapf(err, "urlParam %s is not a valid %T", urlParam, t))
			}

			return i
		case bool:
			i, err := strconv.ParseBool(urlParam)
			if err != nil {
				panic(errors.Wrapf(err, "urlParam %s is not a valid %T", urlParam, t))
			}

			return i
		}

		valType := reflect.TypeOf(val)
		if valType.Kind() == reflect.Pointer {
			val = reflect.New(valType.Elem()).Interface().(T)
			if resolveInterface(urlParam, val) {
				return val
			}
		} else if resolveInterface(urlParam, &val) {
			return val
		}

		panic(fmt.Sprintf("support for %T has not been implemented", val))
	}

	v := fetchParam()
	val, ok := v.(T)
	if !ok {
		panic(fmt.Sprintf("implementation error: returned %T instead of %T", v, val))
	}

	return val
}

func resolveInterface(urlParam string, val any) bool {
	switch t := val.(type) {
	case encoding.TextUnmarshaler:
		if err := t.UnmarshalText([]byte(urlParam)); err != nil {
			panic(errors.Wrap(err, "encoding.UnmarshalText()"))
		}

		return true
	}

	return false
}
