package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestValidateHashEmptyHeader verifies that a missing HashSHA256 header is rejected with 400.
func TestValidateHashEmptyHeader(t *testing.T) {
	t.Parallel()
	h := &MetricsHandler{hashKey: testHashKey, logger: zap.NewNop()}
	handler := h.validateHashMiddleware(fixedResponseHandler("application/json; charset=utf-8", ""))
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("HashSHA256", "")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestValidateHashGetHeadRequestSkipped verifies that GET and HEAD requests without HashSHA256 reach the handler (not 400) when the key is enabled.
func TestValidateHashGetHeadRequestSkipped(t *testing.T) {
	t.Parallel()
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			h := &MetricsHandler{hashKey: testHashKey, logger: zap.NewNop()}
			handler := h.validateHashMiddleware(fixedResponseHandler("application/json; charset=utf-8", `{"status":"ok"}`))
			req := httptest.NewRequest(method, "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
		})
	}
}

// TestWriteHashHeaderGetHeadResponseNotSigned verifies that GET and HEAD responses are delivered without the HashSHA256 header.
func TestWriteHashHeaderGetHeadResponseNotSigned(t *testing.T) {
	t.Parallel()
	body := `{"id":"test","type":"gauge","value":1}`
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			h := &MetricsHandler{hashKey: testHashKey, logger: zap.NewNop()}
			handler := h.writeHashHeaderMiddleware(fixedResponseHandler("application/json; charset=utf-8", body))
			req := httptest.NewRequest(method, "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, body, rec.Body.String())
			assert.Empty(t, rec.Header().Get("HashSHA256"))
		})
	}
}

// TestValidateHashCorrectHeaderValue verifies that a valid hash lets the request through with its body intact.
func TestValidateHashCorrectHeaderValue(t *testing.T) {
	t.Parallel()
	h := &MetricsHandler{hashKey: testHashKey, logger: zap.NewNop()}

	var gotBody []byte
	var err error

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatal("read body:", err)
		}
	})

	handler := h.validateHashMiddleware(inner)

	body := `{"id":"test","type":"gauge","value":1}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	hasher := hmac.New(sha256.New, []byte(h.hashKey))
	hasher.Write([]byte(body))
	calculatedHash := hasher.Sum(nil)
	calculatedHashHex := hex.EncodeToString(calculatedHash)

	req.Header.Set("HashSHA256", calculatedHashHex)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.Equal(t, body, string(gotBody))
}

// TestValidateHashWrongHeaderValue verifies that a mismatching HashSHA256 header is rejected with 400.
func TestValidateHashWrongHeaderValue(t *testing.T) {
	t.Parallel()
	h := &MetricsHandler{hashKey: testHashKey, logger: zap.NewNop()}
	handler := h.validateHashMiddleware(fixedResponseHandler("application/json; charset=utf-8", ""))
	body := `{"id":"test","type":"gauge","value":1}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	hasher := hmac.New(sha256.New, []byte("WRONG VALUE"))
	hasher.Write([]byte(body))
	calculatedHash := hasher.Sum(nil)
	calculatedHashHex := hex.EncodeToString(calculatedHash)

	req.Header.Set("HashSHA256", calculatedHashHex)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestWriteHashHeaderOK verifies that a successful response is signed and delivered unchanged.
func TestWriteHashHeaderOK(t *testing.T) {
	t.Parallel()
	h := &MetricsHandler{hashKey: testHashKey, logger: zap.NewNop()}
	body := `{"id":"test","type":"gauge","value":1}`
	handler := h.writeHashHeaderMiddleware(fixedResponseHandler("application/json; charset=utf-8", body))
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	hasher := hmac.New(sha256.New, []byte(h.hashKey))
	hasher.Write([]byte(body))
	calculatedHash := hasher.Sum(nil)
	calculatedHashHex := hex.EncodeToString(calculatedHash)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, body, rec.Body.String())
	assert.Equal(t, calculatedHashHex, rec.Header().Get("HashSHA256"))
}

// TestWriteHashHeaderNonOKStatus verifies that a non-200 response is still signed and its status preserved.
func TestWriteHashHeaderNonOKStatus(t *testing.T) {
	t.Parallel()
	h := &MetricsHandler{hashKey: testHashKey, logger: zap.NewNop()}
	body := `{"id":"test","type":"gauge","value":1}`
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(body))
	})

	handler := h.writeHashHeaderMiddleware(inner)

	hasher := hmac.New(sha256.New, []byte(h.hashKey))
	hasher.Write([]byte(body))
	calculatedHash := hasher.Sum(nil)
	calculatedHashHex := hex.EncodeToString(calculatedHash)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, body, rec.Body.String())
	assert.Equal(t, calculatedHashHex, rec.Header().Get("HashSHA256"))
}
