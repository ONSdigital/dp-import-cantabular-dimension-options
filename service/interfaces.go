package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
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
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
	Checker(context.Context, *healthcheck.CheckState) error
}

// DatasetAPIClient wraps the DatasetAPIClient handler interface, adding the Checker method
type DatasetAPIClient interface {
	PostInstanceDimensions(ctx context.Context, serviceAuthToken, instanceID string, data dataset.OptionPost) error
	Checker(context.Context, *healthcheck.CheckState) error
}
