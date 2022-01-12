// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/service"
	"sync"
)

// Ensure, that CantabularClientMock does implement service.CantabularClient.
// If this is not the case, regenerate this file with moq.
var _ service.CantabularClient = &CantabularClientMock{}

// CantabularClientMock is a mock implementation of service.CantabularClient.
//
// 	func TestSomethingThatUsesCantabularClient(t *testing.T) {
//
// 		// make and configure a mocked service.CantabularClient
// 		mockedCantabularClient := &CantabularClientMock{
// 			CheckerFunc: func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
// 				panic("mock out the Checker method")
// 			},
// 			GetCodebookFunc: func(contextMoqParam context.Context, getCodebookRequest cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
// 				panic("mock out the GetCodebook method")
// 			},
// 		}
//
// 		// use mockedCantabularClient in code that requires service.CantabularClient
// 		// and then make assertions.
//
// 	}
type CantabularClientMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error

	// GetCodebookFunc mocks the GetCodebook method.
	GetCodebookFunc func(contextMoqParam context.Context, getCodebookRequest cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// CheckState is the checkState argument value.
			CheckState *healthcheck.CheckState
		}
		// GetCodebook holds details about calls to the GetCodebook method.
		GetCodebook []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// GetCodebookRequest is the getCodebookRequest argument value.
			GetCodebookRequest cantabular.GetCodebookRequest
		}
	}
	lockChecker     sync.RWMutex
	lockGetCodebook sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *CantabularClientMock) Checker(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("CantabularClientMock.CheckerFunc: method is nil but CantabularClient.Checker was just called")
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
//     len(mockedCantabularClient.CheckerCalls())
func (mock *CantabularClientMock) CheckerCalls() []struct {
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

// GetCodebook calls GetCodebookFunc.
func (mock *CantabularClientMock) GetCodebook(contextMoqParam context.Context, getCodebookRequest cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
	if mock.GetCodebookFunc == nil {
		panic("CantabularClientMock.GetCodebookFunc: method is nil but CantabularClient.GetCodebook was just called")
	}
	callInfo := struct {
		ContextMoqParam    context.Context
		GetCodebookRequest cantabular.GetCodebookRequest
	}{
		ContextMoqParam:    contextMoqParam,
		GetCodebookRequest: getCodebookRequest,
	}
	mock.lockGetCodebook.Lock()
	mock.calls.GetCodebook = append(mock.calls.GetCodebook, callInfo)
	mock.lockGetCodebook.Unlock()
	return mock.GetCodebookFunc(contextMoqParam, getCodebookRequest)
}

// GetCodebookCalls gets all the calls that were made to GetCodebook.
// Check the length with:
//     len(mockedCantabularClient.GetCodebookCalls())
func (mock *CantabularClientMock) GetCodebookCalls() []struct {
	ContextMoqParam    context.Context
	GetCodebookRequest cantabular.GetCodebookRequest
} {
	var calls []struct {
		ContextMoqParam    context.Context
		GetCodebookRequest cantabular.GetCodebookRequest
	}
	mock.lockGetCodebook.RLock()
	calls = mock.calls.GetCodebook
	mock.lockGetCodebook.RUnlock()
	return calls
}
