package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthCheck.go -pkg mock . HealthChecker
//go:generate moq -out mock/cantabular-client.go -pkg mock . CantabularClient
//go:generate moq -out mock/dataset-api-client.go -pkg mock . DatasetAPIClient
//go:generate moq -out mock/import-api-client.go -pkg mock . ImportAPIClient

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
	AddAndGetCheck(name string, checker healthcheck.Checker) (check *healthcheck.Check, err error)
	Subscribe(s healthcheck.Subscriber, checks ...*healthcheck.Check)
}

// CantabularClient defines the required CantabularClient methods
type CantabularClient interface {
	GetDimensionOptions(context.Context, cantabular.GetDimensionOptionsRequest) (*cantabular.GetDimensionOptionsResponse, error)
	GetAggregatedDimensionOptions(context.Context, cantabular.GetAggregatedDimensionOptionsRequest) (*cantabular.GetAggregatedDimensionOptionsResponse, error)
	Checker(context.Context, *healthcheck.CheckState) error
	CheckerAPIExt(context.Context, *healthcheck.CheckState) error
}

// DatasetAPIClient defines the required Dataset API methods
type DatasetAPIClient interface {
	GetInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID, ifMatch string) (m dataset.Instance, eTag string, err error)
	PatchInstanceDimensions(ctx context.Context, serviceAuthToken string, instanceID string, upserts []*dataset.OptionPost, updates []*dataset.OptionUpdate, ifMatch string) (eTag string, err error)
	PutInstanceState(ctx context.Context, serviceAuthToken, instanceID string, state dataset.State, ifMatch string) (eTag string, err error)
	Checker(context.Context, *healthcheck.CheckState) error
}

// ImportAPIClient defines the required Import API methods
type ImportAPIClient interface {
	UpdateImportJobState(ctx context.Context, jobID string, serviceToken string, newState importapi.State) error
	IncreaseProcessedInstanceCount(ctx context.Context, jobID, serviceToken, instanceID string) (procInst []importapi.ProcessedInstances, err error)
	Checker(context.Context, *healthcheck.CheckState) error
}
