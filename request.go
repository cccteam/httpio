package httpio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
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

func Requests(r *http.Request) iter.Seq2[*http.Request, error] {
	return func(yield func(r *http.Request, err error) bool) {
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
			ctx, err = withParams(ctx, method, op.Path)
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

		for {
			t, err := dec.Token()
			if err != nil {
				yield(nil, NewBadRequestMessageWithErrorf(err, "failed find end of array"))

				return
			}
			token := fmt.Sprintf("%s", t)
			if token == "]" {
				return
			}
			if strings.TrimSpace(token) != "" {
				yield(nil, NewBadRequestMessagef("expected end of array, got %q", t))

				return
			}
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

func withParams(ctx context.Context, method, path string) (context.Context, error) {
	switch method {
	case http.MethodPatch, http.MethodDelete:
		if !strings.HasPrefix(path, "/") {
			return ctx, NewBadRequestMessagef("invalid path %q", path)
		}

		var chiContext chi.Context
		pathParts := strings.Split(path[1:], "/")
		switch {
		case len(pathParts) == 1:
			if pathParts[0] == "" {
				return ctx, NewBadRequestMessagef("invalid path %q", path)
			}
			chiContext.URLParams.Add("id", pathParts[0])
		case len(pathParts) == 2:
			if pathParts[0] == "" || pathParts[1] == "" {
				return ctx, NewBadRequestMessagef("invalid path %q", path)
			}
			chiContext.URLParams.Add("resource", pathParts[0])
			chiContext.URLParams.Add("id", pathParts[1])
		default:
			return ctx, NewBadRequestMessagef("invalid path %q", path)
		}

		ctx = context.WithValue(ctx, chi.RouteCtxKey, &chiContext)
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
