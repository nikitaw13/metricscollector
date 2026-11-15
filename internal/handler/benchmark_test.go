package handler

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkDecompressMiddleware(b *testing.B) {
	var payload bytes.Buffer
	gw := gzip.NewWriter(&payload)
	gw.Write(make([]byte, 1024))
	gw.Close()

	var next http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
	})
	b.ReportAllocs()
	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(payload.Bytes()))
		req.Header.Set("Content-Encoding", "gzip")
		rec := httptest.NewRecorder()
		decompressMiddleware(next).ServeHTTP(rec, req)
	}
}
