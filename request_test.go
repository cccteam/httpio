package httpio

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRequests(t *testing.T) {
	t.Parallel()

	type args struct {
		r *http.Request
	}
	tests := []struct {
		name         string
		args         args
		wantMethod   []string
		wantResource []string
		wantIDs      []string
		wantValues   []string
		wantErr      bool
	}{
		{
			name: "Test Requests with invalid path",
			args: args{
				r: &http.Request{
					Method: "POST",
					Body:   io.NopCloser(bytes.NewBufferString(`[{"op":"patch","path":"/a/b/c","value":{"c":1}}]`)),
				},
			},
			wantErr: true,
		},
		{
			name: "Test Requests with invalid empty path",
			args: args{
				r: &http.Request{
					Method: "POST",
					Body:   io.NopCloser(bytes.NewBufferString(`[{"op":"patch","path":"","value":{"c":1}}]`)),
				},
			},
			wantErr: true,
		},
		{
			name: "Test Requests with invalid id path",
			args: args{
				r: &http.Request{
					Method: "POST",
					Body:   io.NopCloser(bytes.NewBufferString(`[{"op":"patch","path":"/","value":{"c":1}}]`)),
				},
			},
			wantErr: true,
		},
		{
			name: "Test Requests with invalid resource path",
			args: args{
				r: &http.Request{
					Method: "POST",
					Body:   io.NopCloser(bytes.NewBufferString(`[{"op":"patch","path":"/resource/","value":{"c":1}}]`)),
				},
			},
			wantErr: true,
		},
		{
			name: "Test patch Requests with id",
			args: args{
				r: &http.Request{
					Method: "POST",
					Body: io.NopCloser(bytes.NewBufferString(
						`[
							{"op":"patch","path":"/10","value":{"c":1}},
							{"op":"patch","path":"/11","value":{"a":2}}
						]`,
					)),
				},
			},
			wantMethod: []string{"PATCH", "PATCH"},
			wantIDs:    []string{"10", "11"},
			wantValues: []string{`{"c":1}`, `{"a":2}`},
		},
		{
			name: "Test patch Requests with resource and id",
			args: args{
				r: &http.Request{
					Method: "POST",
					Body: io.NopCloser(bytes.NewBufferString(
						`[
							{"op":"patch","path":"/resource1/10","value":{"c":1}},
							{"op":"patch","path":"/resource2/11","value":{"a":2}}
						]`,
					)),
				},
			},
			wantMethod:   []string{"PATCH", "PATCH"},
			wantResource: []string{"resource1", "resource2"},
			wantIDs:      []string{"10", "11"},
			wantValues:   []string{`{"c":1}`, `{"a":2}`},
		},
		{
			name: "Test add Requests with id",
			args: args{
				r: &http.Request{
					Method: "POST",
					Body: io.NopCloser(bytes.NewBufferString(
						`[
							{"op":"add","value":{"c":1}},
							{"op":"add","value":{"a":2}}
						]`,
					)),
				},
			},
			wantMethod: []string{"POST", "POST"},
			wantValues: []string{`{"c":1}`, `{"a":2}`},
		},
		{
			name: "Test delete Requests with id",
			args: args{
				r: &http.Request{
					Method: "POST",
					Body: io.NopCloser(bytes.NewBufferString(
						`[
							{"op":"remove","path":"/10"},
							{"op":"remove","path":"/11"}
						]`,
					)),
				},
			},
			wantMethod: []string{"DELETE", "DELETE"},
			wantIDs:    []string{"10", "11"},
		},
		{
			name: "Test extra space Requests with id",
			args: args{
				r: &http.Request{
					Method: "POST",
					Body: io.NopCloser(bytes.NewBufferString(
						`
							[
								{"op":"add","value":{"c":1}}

							]
						`,
					)),
				},
			},
			wantMethod: []string{"POST"},
			wantValues: []string{`{"c":1}`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotMethod []string
			var gotResource []string
			var gotIDs []string
			var gotValues []string

			for r, err := range Requests(tt.args.r) {
				if (err != nil) != tt.wantErr {
					t.Errorf("Requests() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.wantErr {
					return
				}

				gotMethod = append(gotMethod, r.Method)

				if len(tt.wantResource) > 0 {
					resource := Param[string](r, "resource")
					gotResource = append(gotResource, resource)
				}

				if len(tt.wantIDs) > 0 {
					id := Param[string](r, "id")
					gotIDs = append(gotIDs, id)
				}

				if len(tt.wantValues) > 0 {
					val, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatalf("io.ReadAll() error: %s", err)
					}
					gotValues = append(gotValues, string(val))
				}
			}

			if diff := cmp.Diff(tt.wantMethod, gotMethod); diff != "" {
				t.Errorf("Requests() methods mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantResource, gotResource); diff != "" {
				t.Errorf("Requests() resouces mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantIDs, gotIDs); diff != "" {
				t.Errorf("Requests() IDs mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantValues, gotValues); diff != "" {
				t.Errorf("Requests() values mismatch (-want +got):\n%s", diff)
			}
		})
	}
}