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
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
}

type DatasetAPIClient interface {
	GetInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID, ifMatch string) (m dataset.Instance, eTag string, err error)
	PatchInstanceDimensions(ctx context.Context, serviceAuthToken, instanceID string, data []*dataset.OptionPost, ifMatch string) (eTag string, err error)
	PutInstanceState(ctx context.Context, serviceAuthToken, instanceID string, state dataset.State, ifMatch string) (eTag string, err error)
}

type ImportAPIClient interface {
	UpdateImportJobState(ctx context.Context, jobID, serviceToken string, newState string) error
	IncreaseProcessedInstanceCount(ctx context.Context, jobID, serviceToken, instanceID string) (procInst []importapi.ProcessedInstances, err error)
}
