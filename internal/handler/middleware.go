package handler

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/hmac"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/nikitaw13/metricscollector/internal/sign"
	"go.uber.org/zap"
)

// requireJSONContent enforces application/json Content-Type and non-empty body
// for endpoints that expect JSON payloads.
func requireJSONContent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			http.Error(w, model.ErrMissingType.Error(), http.StatusBadRequest)
			return
		}
		if r.ContentLength == 0 {
			writeJSONError(w, http.StatusBadRequest, errors.New("request body is required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireMetricType rejects requests with an unknown metric type.
func requireMetricType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		if metricType != model.Gauge && metricType != model.Counter {
			http.Error(w, model.ErrInvalidType.Error(), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// responseInfo holds the HTTP status code and response body size.
type responseInfo struct {
	status int
	size   int
}

// loggingResponseWriter wraps http.ResponseWriter to capture status and size.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseInfo *responseInfo
}

// Write records the response body size and defaults status to 200.
func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseInfo.size += size

	if w.responseInfo.status == 0 {
		w.responseInfo.status = http.StatusOK
	}

	return size, err
}

// WriteHeader records the HTTP status code.
func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseInfo.status = statusCode
}

// loggingMiddleware logs incoming HTTP requests and their responses.
func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			respInfo := &responseInfo{
				status: 0,
				size:   0,
			}

			loggingWriter := loggingResponseWriter{
				ResponseWriter: w,
				responseInfo:   respInfo,
			}

			next.ServeHTTP(&loggingWriter, r)

			logger.Info("got incoming HTTP request",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Duration("duration", time.Since(start)),
			)

			logger.Info("response for incoming HTTP request",
				zap.Int("status", respInfo.status),
				zap.Int("size", respInfo.size),
			)
		})
	}
}

// compressibleTypes defines Content-Type values eligible for compression.
var compressibleTypes = map[string]bool{
	"application/json": true,
	"text/html":        true,
}

// compressWriter routes writes through an encoder (gzip/deflate).
// Non-compressible content types bypass the encoder entirely.
type compressWriter struct {
	http.ResponseWriter
	writer   io.WriteCloser // encoder instance
	encoding string         // "gzip" or "deflate"
	bypassed bool           // set when bypassing; prevents encoder Close
}

// WriteHeader checks Content-Type before headers are finalized.
// Sets Content-Encoding only for compressible types; marks non-compressible as bypassed.
func (c *compressWriter) WriteHeader(code int) {
	ct := c.ResponseWriter.Header().Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(ct)
	if compressibleTypes[mediaType] {
		c.ResponseWriter.Header().Set("Content-Encoding", c.encoding)
	} else {
		c.bypassed = true
	}
	c.ResponseWriter.WriteHeader(code)
}

// Write compresses data for eligible Content-Type; otherwise writes directly
// to the underlying ResponseWriter.
func (c *compressWriter) Write(b []byte) (int, error) {
	if c.bypassed {
		return c.ResponseWriter.Write(b)
	}
	ct := c.ResponseWriter.Header().Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(ct)
	if !compressibleTypes[mediaType] {
		c.bypassed = true
		c.ResponseWriter.Header().Del("Content-Encoding")
		return c.ResponseWriter.Write(b)
	}
	if c.ResponseWriter.Header().Get("Content-Encoding") == "" {
		c.ResponseWriter.Header().Set("Content-Encoding", c.encoding)
	}
	return c.writer.Write(b)
}

// compressMiddleware compresses response bodies for clients supporting
// Accept-Encoding. Only application/json and text/html are compressed.
// Gzip takes priority over deflate.
func compressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Accept-Encoding")
		acceptEncoding := r.Header.Get("Accept-Encoding")

		if strings.Contains(acceptEncoding, "gzip") {
			// BestCompression is a package constant, so NewWriterLevel never errors here.
			gw, _ := gzip.NewWriterLevel(w, gzip.BestCompression)
			cw := &compressWriter{ResponseWriter: w, writer: gw, encoding: "gzip"}
			next.ServeHTTP(cw, r)
			if !cw.bypassed {
				gw.Close()
			}
			return
		}

		if strings.Contains(acceptEncoding, "deflate") {
			// BestCompression is a package constant, so NewWriter never errors here.
			dw, _ := flate.NewWriter(w, flate.BestCompression)
			cw := &compressWriter{ResponseWriter: w, writer: dw, encoding: "deflate"}
			next.ServeHTTP(cw, r)
			if !cw.bypassed {
				dw.Close()
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// decompressMiddleware decompresses gzip/deflate request bodies based on
// Content-Encoding. Replaces r.Body and removes encoding/length headers.
func decompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Content-Encoding") {
		case "gzip":
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer gz.Close()
			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length")
			r.ContentLength = -1
			r.Body = io.NopCloser(gz)
		case "deflate":
			df := flate.NewReader(r.Body)
			defer df.Close()
			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length")
			r.ContentLength = -1
			r.Body = io.NopCloser(df)
		}

		next.ServeHTTP(w, r)
	})
}

// validateHashMiddleware rejects requests whose HashSHA256 header does not match
// the HMAC-SHA256 signature of the body computed with the configured key.
// GET and HEAD requests are not validated.
func (h *MetricsHandler) validateHashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GET and HEAD carry no body, so there is nothing to verify.
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		providedHash := []byte(r.Header.Get("HashSHA256"))
		if len(providedHash) == 0 {
			h.logger.Debug("no hash provided")
			http.Error(w, "no hash provided", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error("error reading body", zap.Int("status", http.StatusInternalServerError), zap.Error(err))
			http.Error(w, internalErrorMessage, http.StatusInternalServerError)
			return
		}

		expectedHash := sign.HMAC(h.hashKey, body)

		if !hmac.Equal(providedHash, []byte(expectedHash)) {
			h.logger.Debug("wrong hash provided", zap.String("expectedHash", expectedHash), zap.String("actualHash", fmt.Sprintf("%x", providedHash)))
			http.Error(w, "wrong hash provided", http.StatusBadRequest)
			return
		}
		bodyReader := bytes.NewReader(body)
		r.Body = io.NopCloser(bodyReader)
		next.ServeHTTP(w, r)
	})
}

// hashBodyWriter buffers the response body and status so the signing middleware can compute the body hash after the handler runs.
type hashBodyWriter struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

// Write buffers the body and defaults the status to 200 if no status was set.
func (w *hashBodyWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body.Write(b)
	return len(b), nil
}

// WriteHeader records the response status code.
func (w *hashBodyWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

// writeHashHeaderMiddleware computes the HMAC-SHA256 hash of the response body
// and sends it in the HashSHA256 response header. GET and HEAD responses are
// not signed.
func (h *MetricsHandler) writeHashHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GET and HEAD carry no body, so there is nothing to sign.
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		bw := &hashBodyWriter{
			ResponseWriter: w,
			body:           bytes.NewBuffer(nil),
			status:         http.StatusOK,
		}
		next.ServeHTTP(bw, r)

		body := bw.body.Bytes()

		w.Header().Set("HashSHA256", sign.HMAC(h.hashKey, body))
		w.WriteHeader(bw.status)
		w.Write(body)
	})
}
