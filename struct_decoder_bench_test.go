package httpio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkStructDecoder_Decode(b *testing.B) {
	type request struct {
		Name        string
		Description string
		Count       int
		Tags        []string
		Nested      struct {
			A int
			B string
		}
	}

	benchmarks := []struct {
		name string
		body string
	}{
		{name: "small body", body: `{"Name":"Zach","Description":"a short note","Count":3,"Tags":["a","b","c"],"Nested":{"A":1,"B":"x"}}`},
		{name: "8 KiB body", body: `{"Name":"Zach","Description":"` + strings.Repeat("eight kibibytes of description ", 256) + `","Count":3,"Tags":["a","b","c"],"Nested":{"A":1,"B":"x"}}`},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			decoder, err := NewStructDecoder[request]()
			if err != nil {
				b.Fatalf("NewStructDecoder() error = %v", err)
			}
			ctx := context.Background()

			b.ReportAllocs()
			for b.Loop() {
				r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", strings.NewReader(bm.body))
				if _, err := decoder.Decode(r); err != nil {
					b.Fatalf("StructDecoder.Decode() error = %v", err)
				}
			}
		})
	}
}
