// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/service"
	"sync"
)

// Ensure, that ImportAPIClientMock does implement service.ImportAPIClient.
// If this is not the case, regenerate this file with moq.
var _ service.ImportAPIClient = &ImportAPIClientMock{}

// ImportAPIClientMock is a mock implementation of service.ImportAPIClient.
//
// 	func TestSomethingThatUsesImportAPIClient(t *testing.T) {
//
// 		// make and configure a mocked service.ImportAPIClient
// 		mockedImportAPIClient := &ImportAPIClientMock{
// 			CheckerFunc: func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
// 				panic("mock out the Checker method")
// 			},
// 			IncreaseProcessedInstanceCountFunc: func(ctx context.Context, jobID string, serviceToken string, instanceID string) ([]importapi.ProcessedInstances, error) {
// 				panic("mock out the IncreaseProcessedInstanceCount method")
// 			},
// 			UpdateImportJobStateFunc: func(ctx context.Context, jobID string, serviceToken string, newState importapi.State) error {
// 				panic("mock out the UpdateImportJobState method")
// 			},
// 		}
//
// 		// use mockedImportAPIClient in code that requires service.ImportAPIClient
// 		// and then make assertions.
//
// 	}
type ImportAPIClientMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error

	// IncreaseProcessedInstanceCountFunc mocks the IncreaseProcessedInstanceCount method.
	IncreaseProcessedInstanceCountFunc func(ctx context.Context, jobID string, serviceToken string, instanceID string) ([]importapi.ProcessedInstances, error)

	// UpdateImportJobStateFunc mocks the UpdateImportJobState method.
	UpdateImportJobStateFunc func(ctx context.Context, jobID string, serviceToken string, newState importapi.State) error

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// CheckState is the checkState argument value.
			CheckState *healthcheck.CheckState
		}
		// IncreaseProcessedInstanceCount holds details about calls to the IncreaseProcessedInstanceCount method.
		IncreaseProcessedInstanceCount []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// JobID is the jobID argument value.
			JobID string
			// ServiceToken is the serviceToken argument value.
			ServiceToken string
			// InstanceID is the instanceID argument value.
			InstanceID string
		}
		// UpdateImportJobState holds details about calls to the UpdateImportJobState method.
		UpdateImportJobState []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// JobID is the jobID argument value.
			JobID string
			// ServiceToken is the serviceToken argument value.
			ServiceToken string
			// NewState is the newState argument value.
			NewState importapi.State
		}
	}
	lockChecker                        sync.RWMutex
	lockIncreaseProcessedInstanceCount sync.RWMutex
	lockUpdateImportJobState           sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *ImportAPIClientMock) Checker(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("ImportAPIClientMock.CheckerFunc: method is nil but ImportAPIClient.Checker was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}{
		ContextMoqParam: contextMoqParam,
		CheckState:      checkState,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(contextMoqParam, checkState)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedImportAPIClient.CheckerCalls())
func (mock *ImportAPIClientMock) CheckerCalls() []struct {
	ContextMoqParam context.Context
	CheckState      *healthcheck.CheckState
} {
	var calls []struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// IncreaseProcessedInstanceCount calls IncreaseProcessedInstanceCountFunc.
func (mock *ImportAPIClientMock) IncreaseProcessedInstanceCount(ctx context.Context, jobID string, serviceToken string, instanceID string) ([]importapi.ProcessedInstances, error) {
	if mock.IncreaseProcessedInstanceCountFunc == nil {
		panic("ImportAPIClientMock.IncreaseProcessedInstanceCountFunc: method is nil but ImportAPIClient.IncreaseProcessedInstanceCount was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		JobID        string
		ServiceToken string
		InstanceID   string
	}{
		Ctx:          ctx,
		JobID:        jobID,
		ServiceToken: serviceToken,
		InstanceID:   instanceID,
	}
	mock.lockIncreaseProcessedInstanceCount.Lock()
	mock.calls.IncreaseProcessedInstanceCount = append(mock.calls.IncreaseProcessedInstanceCount, callInfo)
	mock.lockIncreaseProcessedInstanceCount.Unlock()
	return mock.IncreaseProcessedInstanceCountFunc(ctx, jobID, serviceToken, instanceID)
}

// IncreaseProcessedInstanceCountCalls gets all the calls that were made to IncreaseProcessedInstanceCount.
// Check the length with:
//     len(mockedImportAPIClient.IncreaseProcessedInstanceCountCalls())
func (mock *ImportAPIClientMock) IncreaseProcessedInstanceCountCalls() []struct {
	Ctx          context.Context
	JobID        string
	ServiceToken string
	InstanceID   string
} {
	var calls []struct {
		Ctx          context.Context
		JobID        string
		ServiceToken string
		InstanceID   string
	}
	mock.lockIncreaseProcessedInstanceCount.RLock()
	calls = mock.calls.IncreaseProcessedInstanceCount
	mock.lockIncreaseProcessedInstanceCount.RUnlock()
	return calls
}

// UpdateImportJobState calls UpdateImportJobStateFunc.
func (mock *ImportAPIClientMock) UpdateImportJobState(ctx context.Context, jobID string, serviceToken string, newState importapi.State) error {
	if mock.UpdateImportJobStateFunc == nil {
		panic("ImportAPIClientMock.UpdateImportJobStateFunc: method is nil but ImportAPIClient.UpdateImportJobState was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		JobID        string
		ServiceToken string
		NewState     importapi.State
	}{
		Ctx:          ctx,
		JobID:        jobID,
		ServiceToken: serviceToken,
		NewState:     newState,
	}
	mock.lockUpdateImportJobState.Lock()
	mock.calls.UpdateImportJobState = append(mock.calls.UpdateImportJobState, callInfo)
	mock.lockUpdateImportJobState.Unlock()
	return mock.UpdateImportJobStateFunc(ctx, jobID, serviceToken, newState)
}

// UpdateImportJobStateCalls gets all the calls that were made to UpdateImportJobState.
// Check the length with:
//     len(mockedImportAPIClient.UpdateImportJobStateCalls())
func (mock *ImportAPIClientMock) UpdateImportJobStateCalls() []struct {
	Ctx          context.Context
	JobID        string
	ServiceToken string
	NewState     importapi.State
} {
	var calls []struct {
		Ctx          context.Context
		JobID        string
		ServiceToken string
		NewState     importapi.State
	}
	mock.lockUpdateImportJobState.RLock()
	calls = mock.calls.UpdateImportJobState
	mock.lockUpdateImportJobState.RUnlock()
	return calls
}
