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

type OperationType string

const (
	OperationCreate OperationType = "add"
	OperationUpdate OperationType = "patch"
	OperationDelete OperationType = "remove"
)

type Operation struct {
	Type OperationType
	Req  *http.Request
}

type patchOperation struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value"`
}

func Operations(r *http.Request, pattern string) iter.Seq2[*Operation, error] {
	return func(yield func(r *Operation, err error) bool) {
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
			var op patchOperation
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
			ctx = withParams(ctx, method, pattern, op.Path)
			r2, err := http.NewRequestWithContext(ctx, method, op.Path, bytes.NewReader([]byte(op.Value)))
			if err != nil {
				yield(nil, err)

				return
			}

			if !yield(&Operation{Type: OperationType(op.Op), Req: r2}, nil) {
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
	switch OperationType(strings.ToLower(op)) {
	case OperationCreate:
		return http.MethodPost, nil
	case OperationUpdate:
		return http.MethodPatch, nil
	case OperationDelete:
		return http.MethodDelete, nil
	default:
		return "", errors.Newf("unsupported operation %q", op)
	}
}

func withParams(ctx context.Context, method, pattern, path string) context.Context {
	switch method {
	case http.MethodPatch, http.MethodDelete:
		var chiContext *chi.Context
		r := chi.NewRouter()
		r.Handle(pattern, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			chiContext = chi.RouteContext(r.Context())
		}))
		r.ServeHTTP(nil, &http.Request{Method: method, URL: &url.URL{Path: path}})

		ctx = context.WithValue(ctx, chi.RouteCtxKey, chiContext)
	}

	return ctx
}

func permissionFromType(typ OperationType) accesstypes.Permission {
	switch typ {
	case OperationCreate:
		return accesstypes.Create
	case OperationUpdate:
		return accesstypes.Update
	case OperationDelete:
		return accesstypes.Delete
	}

	panic("implementation error")
}
