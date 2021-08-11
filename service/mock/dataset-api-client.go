// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/service"
	"sync"
)

var (
	lockDatasetAPIClientMockChecker                sync.RWMutex
	lockDatasetAPIClientMockGetInstance            sync.RWMutex
	lockDatasetAPIClientMockPostInstanceDimensions sync.RWMutex
	lockDatasetAPIClientMockPutInstanceState       sync.RWMutex
)

// Ensure, that DatasetAPIClientMock does implement service.DatasetAPIClient.
// If this is not the case, regenerate this file with moq.
var _ service.DatasetAPIClient = &DatasetAPIClientMock{}

// DatasetAPIClientMock is a mock implementation of service.DatasetAPIClient.
//
//     func TestSomethingThatUsesDatasetAPIClient(t *testing.T) {
//
//         // make and configure a mocked service.DatasetAPIClient
//         mockedDatasetAPIClient := &DatasetAPIClientMock{
//             CheckerFunc: func(in1 context.Context, in2 *healthcheck.CheckState) error {
// 	               panic("mock out the Checker method")
//             },
//             GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
// 	               panic("mock out the GetInstance method")
//             },
//             PostInstanceDimensionsFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost, ifMatch string) (string, error) {
// 	               panic("mock out the PostInstanceDimensions method")
//             },
//             PutInstanceStateFunc: func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
// 	               panic("mock out the PutInstanceState method")
//             },
//         }
//
//         // use mockedDatasetAPIClient in code that requires service.DatasetAPIClient
//         // and then make assertions.
//
//     }
type DatasetAPIClientMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(in1 context.Context, in2 *healthcheck.CheckState) error

	// GetInstanceFunc mocks the GetInstance method.
	GetInstanceFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error)

	// PostInstanceDimensionsFunc mocks the PostInstanceDimensions method.
	PostInstanceDimensionsFunc func(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost, ifMatch string) (string, error)

	// PutInstanceStateFunc mocks the PutInstanceState method.
	PutInstanceStateFunc func(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// In1 is the in1 argument value.
			In1 context.Context
			// In2 is the in2 argument value.
			In2 *healthcheck.CheckState
		}
		// GetInstance holds details about calls to the GetInstance method.
		GetInstance []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// InstanceID is the instanceID argument value.
			InstanceID string
			// IfMatch is the ifMatch argument value.
			IfMatch string
		}
		// PostInstanceDimensions holds details about calls to the PostInstanceDimensions method.
		PostInstanceDimensions []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// InstanceID is the instanceID argument value.
			InstanceID string
			// Data is the data argument value.
			Data dataset.OptionPost
			// IfMatch is the ifMatch argument value.
			IfMatch string
		}
		// PutInstanceState holds details about calls to the PutInstanceState method.
		PutInstanceState []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// InstanceID is the instanceID argument value.
			InstanceID string
			// State is the state argument value.
			State dataset.State
			// IfMatch is the ifMatch argument value.
			IfMatch string
		}
	}
}

// Checker calls CheckerFunc.
func (mock *DatasetAPIClientMock) Checker(in1 context.Context, in2 *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("DatasetAPIClientMock.CheckerFunc: method is nil but DatasetAPIClient.Checker was just called")
	}
	callInfo := struct {
		In1 context.Context
		In2 *healthcheck.CheckState
	}{
		In1: in1,
		In2: in2,
	}
	lockDatasetAPIClientMockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	lockDatasetAPIClientMockChecker.Unlock()
	return mock.CheckerFunc(in1, in2)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedDatasetAPIClient.CheckerCalls())
func (mock *DatasetAPIClientMock) CheckerCalls() []struct {
	In1 context.Context
	In2 *healthcheck.CheckState
} {
	var calls []struct {
		In1 context.Context
		In2 *healthcheck.CheckState
	}
	lockDatasetAPIClientMockChecker.RLock()
	calls = mock.calls.Checker
	lockDatasetAPIClientMockChecker.RUnlock()
	return calls
}

// GetInstance calls GetInstanceFunc.
func (mock *DatasetAPIClientMock) GetInstance(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
	if mock.GetInstanceFunc == nil {
		panic("DatasetAPIClientMock.GetInstanceFunc: method is nil but DatasetAPIClient.GetInstance was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		InstanceID       string
		IfMatch          string
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		InstanceID:       instanceID,
		IfMatch:          ifMatch,
	}
	lockDatasetAPIClientMockGetInstance.Lock()
	mock.calls.GetInstance = append(mock.calls.GetInstance, callInfo)
	lockDatasetAPIClientMockGetInstance.Unlock()
	return mock.GetInstanceFunc(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, ifMatch)
}

// GetInstanceCalls gets all the calls that were made to GetInstance.
// Check the length with:
//     len(mockedDatasetAPIClient.GetInstanceCalls())
func (mock *DatasetAPIClientMock) GetInstanceCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	InstanceID       string
	IfMatch          string
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		InstanceID       string
		IfMatch          string
	}
	lockDatasetAPIClientMockGetInstance.RLock()
	calls = mock.calls.GetInstance
	lockDatasetAPIClientMockGetInstance.RUnlock()
	return calls
}

// PostInstanceDimensions calls PostInstanceDimensionsFunc.
func (mock *DatasetAPIClientMock) PostInstanceDimensions(ctx context.Context, serviceAuthToken string, instanceID string, data dataset.OptionPost, ifMatch string) (string, error) {
	if mock.PostInstanceDimensionsFunc == nil {
		panic("DatasetAPIClientMock.PostInstanceDimensionsFunc: method is nil but DatasetAPIClient.PostInstanceDimensions was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		ServiceAuthToken string
		InstanceID       string
		Data             dataset.OptionPost
		IfMatch          string
	}{
		Ctx:              ctx,
		ServiceAuthToken: serviceAuthToken,
		InstanceID:       instanceID,
		Data:             data,
		IfMatch:          ifMatch,
	}
	lockDatasetAPIClientMockPostInstanceDimensions.Lock()
	mock.calls.PostInstanceDimensions = append(mock.calls.PostInstanceDimensions, callInfo)
	lockDatasetAPIClientMockPostInstanceDimensions.Unlock()
	return mock.PostInstanceDimensionsFunc(ctx, serviceAuthToken, instanceID, data, ifMatch)
}

// PostInstanceDimensionsCalls gets all the calls that were made to PostInstanceDimensions.
// Check the length with:
//     len(mockedDatasetAPIClient.PostInstanceDimensionsCalls())
func (mock *DatasetAPIClientMock) PostInstanceDimensionsCalls() []struct {
	Ctx              context.Context
	ServiceAuthToken string
	InstanceID       string
	Data             dataset.OptionPost
	IfMatch          string
} {
	var calls []struct {
		Ctx              context.Context
		ServiceAuthToken string
		InstanceID       string
		Data             dataset.OptionPost
		IfMatch          string
	}
	lockDatasetAPIClientMockPostInstanceDimensions.RLock()
	calls = mock.calls.PostInstanceDimensions
	lockDatasetAPIClientMockPostInstanceDimensions.RUnlock()
	return calls
}

// PutInstanceState calls PutInstanceStateFunc.
func (mock *DatasetAPIClientMock) PutInstanceState(ctx context.Context, serviceAuthToken string, instanceID string, state dataset.State, ifMatch string) (string, error) {
	if mock.PutInstanceStateFunc == nil {
		panic("DatasetAPIClientMock.PutInstanceStateFunc: method is nil but DatasetAPIClient.PutInstanceState was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		ServiceAuthToken string
		InstanceID       string
		State            dataset.State
		IfMatch          string
	}{
		Ctx:              ctx,
		ServiceAuthToken: serviceAuthToken,
		InstanceID:       instanceID,
		State:            state,
		IfMatch:          ifMatch,
	}
	lockDatasetAPIClientMockPutInstanceState.Lock()
	mock.calls.PutInstanceState = append(mock.calls.PutInstanceState, callInfo)
	lockDatasetAPIClientMockPutInstanceState.Unlock()
	return mock.PutInstanceStateFunc(ctx, serviceAuthToken, instanceID, state, ifMatch)
}

// PutInstanceStateCalls gets all the calls that were made to PutInstanceState.
// Check the length with:
//     len(mockedDatasetAPIClient.PutInstanceStateCalls())
func (mock *DatasetAPIClientMock) PutInstanceStateCalls() []struct {
	Ctx              context.Context
	ServiceAuthToken string
	InstanceID       string
	State            dataset.State
	IfMatch          string
} {
	var calls []struct {
		Ctx              context.Context
		ServiceAuthToken string
		InstanceID       string
		State            dataset.State
		IfMatch          string
	}
	lockDatasetAPIClientMockPutInstanceState.RLock()
	calls = mock.calls.PutInstanceState
	lockDatasetAPIClientMockPutInstanceState.RUnlock()
	return calls
}
