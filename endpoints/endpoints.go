package endpoints

import (
	"net/http"

	"github.com/Moranilt/config-keeper/config/database"
	"github.com/Moranilt/config-keeper/healthcheck"
	"github.com/Moranilt/config-keeper/middleware"
	"github.com/Moranilt/config-keeper/service"
)

type Endpoint struct {
	Pattern    string
	HandleFunc http.HandlerFunc
	Methods    []string
	Middleware []middleware.EndpointMiddlewareFunc
}

func MakeEndpoints(service service.Service, mw *middleware.Middleware) []Endpoint {
	return []Endpoint{
		{
			Pattern:    "/user",
			HandleFunc: service.CreateUser,
			Methods:    []string{http.MethodPost},
		},
	}
}

func MakeHealth(db *database.Checker) Endpoint {
	return Endpoint{
		Pattern: "/health",
		HandleFunc: healthcheck.HandlerFunc(
			healthcheck.HealthItem{
				Name:    "database",
				Checker: db,
			},
		),
		Methods: []string{http.MethodGet},
	}
}
