package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/handler"
)

//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthCheck.go -pkg mock . HealthChecker
//go:generate moq -out mock/cantabular-client.go -pkg mock . CantabularClient
//go:generate moq -out mock/dataset-api-client.go -pkg mock . DatasetAPIClient

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

// CantabularClient wraps the CantabularClient handler interface, adding the Checker method
type CantabularClient interface {
	handler.CantabularClient
	Checker(context.Context, *healthcheck.CheckState) error
}

// DatasetAPIClient wraps the DatasetAPIClient handler interface, adding the Checker method
type DatasetAPIClient interface {
	handler.DatasetAPIClient
	Checker(context.Context, *healthcheck.CheckState) error
}
