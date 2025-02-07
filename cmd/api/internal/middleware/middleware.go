package middleware

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"telegram/internal/logger"
	"time"
)

type Crypter struct {
	key []byte
}

func NewCrypter(key []byte) *Crypter {
	return &Crypter{
		key: key,
	}
}

func WithJSONContent(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerContentType := r.Header.Get("Content-Type")
		if headerContentType != "application/json" {
			resp := make(map[string]string)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnsupportedMediaType)
			resp["message"] = "Content Type is not application/json"
			jsonResp, _ := json.Marshal(resp)
			_, _ = w.Write(jsonResp)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func WithLogging(h http.Handler, l *logger.Logger) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			l.ErrorCtx(r.Context(), "error reading body", logger.Field{Key: "error", Value: err})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(strings.NewReader(string(b)))

		h.ServeHTTP(w, r)

		duration := time.Since(start)

		l.InfoCtx(r.Context(), "new request",
			logger.Field{Key: "uri", Value: uri},
			logger.Field{Key: "method", Value: method},
			logger.Field{Key: "duration", Value: duration},
			logger.Field{Key: "body", Value: b},
		)
	}

	return http.HandlerFunc(logFn)
}

func WithCompressing(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding := r.Header.Get("Content-Encoding")
		if encoding == "gzip" {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer reader.Close()

			r.Body = io.NopCloser(reader)
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			h.ServeHTTP(w, r)
			return
		}
		defer gz.Close()

		gzrw := gzipResponseWriter{Writer: gz, ResponseWriter: w}

		h.ServeHTTP(gzrw, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	return n, err
}
