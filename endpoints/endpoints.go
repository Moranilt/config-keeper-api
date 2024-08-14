package endpoints

import (
	"net/http"

	"github.com/Moranilt/config-keeper/healthcheck"
	"github.com/Moranilt/config-keeper/middleware"
	"github.com/Moranilt/config-keeper/service"
	"github.com/Moranilt/http-utils/clients/database"

	"net/http/pprof"
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
			Pattern:    "/debug/pprof/profile",
			HandleFunc: pprof.Profile,
			Methods:    []string{http.MethodGet},
		},
		{
			Pattern:    "/folders",
			HandleFunc: service.CreateFolder,
			Methods:    []string{http.MethodPost},
		},
		{
			Pattern:    "/folders/{id}",
			HandleFunc: service.GetFolder,
			Methods:    []string{http.MethodGet},
		},
		{
			Pattern:    "/folders/{id}",
			HandleFunc: service.DeleteFolder,
			Methods:    []string{http.MethodDelete},
		},
		{
			Pattern:    "/folders/{id}",
			HandleFunc: service.EditFolder,
			Methods:    []string{http.MethodPatch},
		},
		{
			Pattern:    "/files",
			HandleFunc: service.CreateFile,
			Methods:    []string{http.MethodPost},
		},
		{
			Pattern:    "/files/{id}",
			HandleFunc: service.DeleteFile,
			Methods:    []string{http.MethodDelete},
		},
		{
			Pattern:    "/files/{id}",
			HandleFunc: service.EditFile,
			Methods:    []string{http.MethodPatch},
		},
		{
			Pattern:    "/files/{id}",
			HandleFunc: service.GetFile,
			Methods:    []string{http.MethodGet},
		},
		{
			Pattern:    "/files/{file_id}/contents",
			HandleFunc: service.GetFileContents,
			Methods:    []string{http.MethodGet},
		},
		{
			Pattern:    "/files/{file_id}/contents",
			HandleFunc: service.CreateFileContent,
			Methods:    []string{http.MethodPost},
		},
		{
			Pattern:    "/files/{file_id}/contents/{id}",
			HandleFunc: service.EditFileContent,
			Methods:    []string{http.MethodPatch},
		},
		{
			Pattern:    "/files/{file_id}/contents/{id}",
			HandleFunc: service.DeleteFileContent,
			Methods:    []string{http.MethodDelete},
		},
		{
			Pattern:    "/files/{file_id}/listeners",
			HandleFunc: service.GetFileListeners,
			Methods:    []string{http.MethodGet},
		},
		{
			Pattern:    "/files/{file_id}/listeners",
			HandleFunc: service.CreateListener,
			Methods:    []string{http.MethodPost},
		},
		{
			Pattern:    "/files/{file_id}/listeners/{id}",
			HandleFunc: service.GetListener,
			Methods:    []string{http.MethodGet},
		},
		{
			Pattern:    "/files/{file_id}/listeners/{id}",
			HandleFunc: service.EditListener,
			Methods:    []string{http.MethodPatch},
		},
		{
			Pattern:    "/files/{file_id}/listeners/{id}",
			HandleFunc: service.DeleteListener,
			Methods:    []string{http.MethodDelete},
		},
	}
}

func MakeHealth(db *database.Client) Endpoint {
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
