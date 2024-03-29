package handler

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
)

//go:generate moq -out mock/cantabular-client.go -pkg mock . CantabularClient
//go:generate moq -out mock/dataset-api-client.go -pkg mock . DatasetAPIClient
//go:generate moq -out mock/import-api-client.go -pkg mock . ImportAPIClient

type CantabularClient interface {
	GetDimensionOptions(context.Context, cantabular.GetDimensionOptionsRequest) (*cantabular.GetDimensionOptionsResponse, error)
	GetAggregatedDimensionOptions(context.Context, cantabular.GetAggregatedDimensionOptionsRequest) (*cantabular.GetAggregatedDimensionOptionsResponse, error)
}

type DatasetAPIClient interface {
	GetInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID, ifMatch string) (m dataset.Instance, eTag string, err error)
	PatchInstanceDimensions(ctx context.Context, serviceAuthToken, instanceID string, upserts []*dataset.OptionPost, updates []*dataset.OptionUpdate, ifMatch string) (eTag string, err error)
	PutInstanceState(ctx context.Context, serviceAuthToken, instanceID string, state dataset.State, ifMatch string) (eTag string, err error)
}

type ImportAPIClient interface {
	UpdateImportJobState(ctx context.Context, jobID, serviceToken string, newState importapi.State) error
	IncreaseProcessedInstanceCount(ctx context.Context, jobID, serviceToken, instanceID string) (procInst []importapi.ProcessedInstances, err error)
}
