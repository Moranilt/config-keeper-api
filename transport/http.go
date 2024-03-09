package transport

import (
	"net/http"
	"time"

	"github.com/Moranilt/config-keeper/endpoints"
	"github.com/Moranilt/config-keeper/middleware"
	"github.com/gorilla/mux"
)

func New(addr string, endpoints []endpoints.Endpoint, mw *middleware.Middleware) *http.Server {
	router := mux.NewRouter()
	router.Use(mw.Default, mw.Otel)

	for _, endpoint := range endpoints {
		handler := applyMiddleware(endpoint.HandleFunc, endpoint.Middleware)
		router.Handle(endpoint.Pattern, handler).Methods(endpoint.Methods...)
	}

	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  40 * time.Second,
		WriteTimeout: 40 * time.Second,
		Handler:      router,
	}
	return server
}

func applyMiddleware(handler http.Handler, mws []middleware.EndpointMiddlewareFunc) http.Handler {
	for _, mw := range mws {
		handler = mw(handler)
	}
	return handler
}
