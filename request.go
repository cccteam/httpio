package httpio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strings"

	"github.com/cccteam/ccc/accesstypes"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/errors/v5"
)

type Operation struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value"`
}

func Requests(r *http.Request, pattern string) iter.Seq2[*http.Request, error] {
	return func(yield func(r *http.Request, err error) bool) {
		if !strings.HasPrefix(pattern, "/") {
			yield(nil, errors.New("pattern must start with /"))

			return
		}

		dec := json.NewDecoder(r.Body)

		for {
			t, err := dec.Token()
			if err != nil {
				yield(nil, err)

				return
			}
			token := fmt.Sprintf("%s", t)
			if token == "[" {
				break
			}
			if strings.TrimSpace(token) != "" {
				yield(nil, NewBadRequestMessagef("expected start of array, got %q", t))

				return
			}
		}

		for dec.More() {
			var op Operation
			if err := dec.Decode(&op); err != nil {
				yield(nil, err)

				return
			}

			method, err := httpMethod(op.Op)
			if err != nil {
				yield(nil, err)

				return
			}

			ctx := r.Context()
			ctx, err = withParams(ctx, method, pattern, op.Path)
			if err != nil {
				yield(nil, err)

				return
			}

			r2, err := http.NewRequestWithContext(ctx, method, op.Path, bytes.NewReader([]byte(op.Value)))
			if err != nil {
				yield(nil, err)

				return
			}

			if !yield(r2, nil) {
				return
			}
		}

		t, err := dec.Token()
		if err != nil {
			yield(nil, NewBadRequestMessageWithErrorf(err, "failed find end of array"))

			return
		}

		token := fmt.Sprintf("%s", t)
		if token == "]" {
			return
		}
	}
}

func httpMethod(op string) (string, error) {
	switch strings.ToLower(op) {
	case "add":
		return http.MethodPost, nil
	case "patch":
		return http.MethodPatch, nil
	case "remove":
		return http.MethodDelete, nil
	default:
		return "", errors.Newf("unsupported operation %q", op)
	}
}

func withParams(ctx context.Context, method, pattern, path string) (context.Context, error) {
	switch method {
	case http.MethodPatch, http.MethodDelete:
		var chiContext *chi.Context
		r := chi.NewRouter()
		r.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			chiContext = chi.RouteContext(r.Context())
		}))
		r.ServeHTTP(nil, &http.Request{Method: method, URL: &url.URL{Path: path}})

		ctx = context.WithValue(ctx, chi.RouteCtxKey, chiContext)
	}

	return ctx, nil
}

func PermissionFromRequest(r *http.Request) (accesstypes.Permission, error) {
	var perm accesstypes.Permission
	switch r.Method {
	case http.MethodPost:
		perm = accesstypes.Create
	case http.MethodPatch:
		perm = accesstypes.Update
	case http.MethodDelete:
		perm = accesstypes.Delete
	default:
		return perm, NewBadRequestWithError(errors.Newf("unsupported method, %s", r.Method))
	}

	return perm, nil
}
