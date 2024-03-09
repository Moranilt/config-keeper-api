package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Moranilt/http-utils/logger"
	"github.com/Moranilt/http-utils/response"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type ContextKey string

const (
	TOKEN_HEADER = "X-App-Token"
)

type Middleware struct {
	logger logger.Logger
}

type EndpointMiddlewareFunc func(handleFunc http.Handler) http.Handler

func New(l logger.Logger) *Middleware {
	return &Middleware{
		logger: l,
	}
}

func (m *Middleware) Default(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := uuid.NewString()
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Request-ID", requestId)
		ctx := context.WithValue(r.Context(), logger.CtxRequestId, requestId)
		r = r.WithContext(ctx)
		m.logger.WithRequestInfo(r).Info("Incoming request")
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) Otel(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		newCtx, span := otel.Tracer("http").Start(ctx, r.URL.Path)
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", r.URL.Path),
			attribute.String("span.kind", "server"),
			attribute.String("request_id", GetRequestID(r.Context())),
		)

		rw := &responseWriter{ResponseWriter: w}
		r = r.WithContext(newCtx)
		next.ServeHTTP(rw, r)

		span.SetAttributes(attribute.Int("http.status_code", rw.statusCode))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (m *Middleware) AppTokenRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(TOKEN_HEADER) == "" {
			response.ErrorResponse(w, fmt.Errorf("%s required", TOKEN_HEADER), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetRequestID(ctx context.Context) string {
	return ctx.Value(logger.CtxRequestId).(string)
}
