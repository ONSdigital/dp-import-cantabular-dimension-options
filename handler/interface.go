package handler

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
)

//go:generate moq -out mock/cantabular-client.go -pkg mock . CantabularClient
//go:generate moq -out mock/dataset-api-client.go -pkg mock . DatasetAPIClient

type CantabularClient interface {
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
}

type DatasetAPIClient interface {
	GetInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID, ifMatch string) (m dataset.Instance, eTag string, err error)
	GetInstanceDimensions(ctx context.Context, serviceAuthToken, instanceID string, q *dataset.QueryParams, ifMatch string) (m dataset.Dimensions, eTag string, err error)
	PostInstanceDimensions(ctx context.Context, serviceAuthToken, instanceID string, data dataset.OptionPost, ifMatch string) (eTag string, err error)
	PutInstanceState(ctx context.Context, serviceAuthToken, instanceID string, state dataset.State, ifMatch string) (eTag string, err error)
}
