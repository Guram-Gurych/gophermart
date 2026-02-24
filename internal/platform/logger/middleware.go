package logger

import (
	"context"
	"github.com/Guram-Gurych/gophermart.git/internal/domain"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

type ReqIDHandler struct {
	slog.Handler
}

func (h *ReqIDHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := ctx.Value(domain.RequestIDKey).(string); ok {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func LoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			respData := &responseData{}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   respData,
			}

			start := time.Now()
			next.ServeHTTP(&lw, r)

			logger.InfoContext(r.Context(), "request processed",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Duration("latency", time.Since(start)),
				slog.Int("status", respData.status),
				slog.Int("size", respData.size),
			)
		})
	}
}

func RequestMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := uuid.New().String()
			w.Header().Set("X-Request-ID", id)

			ctx := context.WithValue(r.Context(), domain.RequestIDKey, id)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
