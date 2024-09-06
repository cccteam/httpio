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
func Param[T any](r *http.Request, param ParamType) T {
	// Fetch the URL parameter
	urlParam := chi.URLParam(r, string(param))
	if urlParam == "" {
		panic(newParamErrMsg("route parameter (%s) is required", param))
	}

	var val T
	v, err := resolve(urlParam, val)
	if err != nil {
		panic(err)
	}

	if v == nil {
		valType := reflect.TypeOf(val)
		if valType.Kind() == reflect.Pointer {
			val = initType(valType).(T)
			if err := resolveInterface(urlParam, val); err != nil {
				panic(err)
			}
		} else {
			if err := resolveInterface(urlParam, &val); err != nil {
				panic(err)
			}
		}
		v = val
	}

	// Type assertion to return value
	val, ok := v.(T)
	if !ok {
		panic(fmt.Sprintf("implementation error: returned %T instead of %T", v, val))
	}

	return val
}

func initType(v reflect.Type) any {
	return reflect.New(v.Elem()).Interface()
}

func resolve(urlParam string, val any) (any, error) {
	switch t := val.(type) {
	case string:
		return urlParam, nil
	case int:
		i, err := strconv.Atoi(urlParam)
		if err != nil {
			return nil, errors.Wrapf(err, "urlParam %s is not a valid %T", urlParam, t)
		}

		return i, nil
	case int64:
		i, err := strconv.ParseInt(urlParam, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "urlParam %s is not a valid %T", urlParam, t)
		}

		return i, nil
	case float64:
		i, err := strconv.ParseFloat(urlParam, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "urlParam %s is not a valid %T", urlParam, t)
		}

		return i, nil
	case bool:
		i, err := strconv.ParseBool(urlParam)
		if err != nil {
			return nil, errors.Wrapf(err, "urlParam %s is not a valid %T", urlParam, t)
		}

		return i, nil
	}

	return nil, nil
}

func resolveInterface(urlParam string, val any) error {
	enc, ok := val.(encoding.TextUnmarshaler)
	if !ok {
		return errors.Newf("unsupported type %T", val)
	}

	if err := enc.UnmarshalText([]byte(urlParam)); err != nil {
		return errors.Wrap(err, "encoding.UnmarshalText()")
	}

	return nil
}
