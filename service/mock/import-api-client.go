// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/service"
	"sync"
)

var (
	lockImportAPIClientMockChecker              sync.RWMutex
	lockImportAPIClientMockUpdateImportJobState sync.RWMutex
)

// Ensure, that ImportAPIClientMock does implement service.ImportAPIClient.
// If this is not the case, regenerate this file with moq.
var _ service.ImportAPIClient = &ImportAPIClientMock{}

// ImportAPIClientMock is a mock implementation of service.ImportAPIClient.
//
//     func TestSomethingThatUsesImportAPIClient(t *testing.T) {
//
//         // make and configure a mocked service.ImportAPIClient
//         mockedImportAPIClient := &ImportAPIClientMock{
//             CheckerFunc: func(in1 context.Context, in2 *healthcheck.CheckState) error {
// 	               panic("mock out the Checker method")
//             },
//             UpdateImportJobStateFunc: func(ctx context.Context, jobID string, serviceToken string, newState string) error {
// 	               panic("mock out the UpdateImportJobState method")
//             },
//         }
//
//         // use mockedImportAPIClient in code that requires service.ImportAPIClient
//         // and then make assertions.
//
//     }
type ImportAPIClientMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(in1 context.Context, in2 *healthcheck.CheckState) error

	// UpdateImportJobStateFunc mocks the UpdateImportJobState method.
	UpdateImportJobStateFunc func(ctx context.Context, jobID string, serviceToken string, newState string) error

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// In1 is the in1 argument value.
			In1 context.Context
			// In2 is the in2 argument value.
			In2 *healthcheck.CheckState
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
			NewState string
		}
	}
}

// Checker calls CheckerFunc.
func (mock *ImportAPIClientMock) Checker(in1 context.Context, in2 *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("ImportAPIClientMock.CheckerFunc: method is nil but ImportAPIClient.Checker was just called")
	}
	callInfo := struct {
		In1 context.Context
		In2 *healthcheck.CheckState
	}{
		In1: in1,
		In2: in2,
	}
	lockImportAPIClientMockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	lockImportAPIClientMockChecker.Unlock()
	return mock.CheckerFunc(in1, in2)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedImportAPIClient.CheckerCalls())
func (mock *ImportAPIClientMock) CheckerCalls() []struct {
	In1 context.Context
	In2 *healthcheck.CheckState
} {
	var calls []struct {
		In1 context.Context
		In2 *healthcheck.CheckState
	}
	lockImportAPIClientMockChecker.RLock()
	calls = mock.calls.Checker
	lockImportAPIClientMockChecker.RUnlock()
	return calls
}

// UpdateImportJobState calls UpdateImportJobStateFunc.
func (mock *ImportAPIClientMock) UpdateImportJobState(ctx context.Context, jobID string, serviceToken string, newState string) error {
	if mock.UpdateImportJobStateFunc == nil {
		panic("ImportAPIClientMock.UpdateImportJobStateFunc: method is nil but ImportAPIClient.UpdateImportJobState was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		JobID        string
		ServiceToken string
		NewState     string
	}{
		Ctx:          ctx,
		JobID:        jobID,
		ServiceToken: serviceToken,
		NewState:     newState,
	}
	lockImportAPIClientMockUpdateImportJobState.Lock()
	mock.calls.UpdateImportJobState = append(mock.calls.UpdateImportJobState, callInfo)
	lockImportAPIClientMockUpdateImportJobState.Unlock()
	return mock.UpdateImportJobStateFunc(ctx, jobID, serviceToken, newState)
}

// UpdateImportJobStateCalls gets all the calls that were made to UpdateImportJobState.
// Check the length with:
//     len(mockedImportAPIClient.UpdateImportJobStateCalls())
func (mock *ImportAPIClientMock) UpdateImportJobStateCalls() []struct {
	Ctx          context.Context
	JobID        string
	ServiceToken string
	NewState     string
} {
	var calls []struct {
		Ctx          context.Context
		JobID        string
		ServiceToken string
		NewState     string
	}
	lockImportAPIClientMockUpdateImportJobState.RLock()
	calls = mock.calls.UpdateImportJobState
	lockImportAPIClientMockUpdateImportJobState.RUnlock()
	return calls
}
