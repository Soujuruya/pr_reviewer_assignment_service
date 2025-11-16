package middleware

import (
	"bytes"
	"io"
	"net/http"

	"pr_reviewer_assignment_service/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func LoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = uuid.NewString()
			}

			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = uuid.NewString()
			}

			ctx := r.Context()
			ctx = logger.WithRequestID(ctx, reqID)
			ctx = logger.WithTraceID(ctx, traceID)
			r = r.WithContext(ctx)

			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			log.Info(ctx, "incoming request",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.ByteString("body", bodyBytes),
			)

			lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK, body: &bytes.Buffer{}}

			next.ServeHTTP(lrw, r)

			log.Info(ctx, "response",
				zap.Int("status", lrw.statusCode),
				zap.ByteString("body", lrw.body.Bytes()),
			)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (l *loggingResponseWriter) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}

func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	l.body.Write(b)
	return l.ResponseWriter.Write(b)
}
