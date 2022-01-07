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
// 			CheckerAPIExtFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
// 				panic("mock out the CheckerAPIExt method")
// 			},
// 			GetDimensionOptionsFunc: func(ctx context.Context, req cantabular.StaticDatasetQueryRequest) (*cantabular.GetDimensionOptionsResponse, error) {
// 				panic("mock out the GetDimensionOptions method")
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

	// CheckerAPIExtFunc mocks the CheckerAPIExt method.
	CheckerAPIExtFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// GetDimensionOptionsFunc mocks the GetDimensionOptions method.
	GetDimensionOptionsFunc func(ctx context.Context, req cantabular.StaticDatasetQueryRequest) (*cantabular.GetDimensionOptionsResponse, error)

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// CheckState is the checkState argument value.
			CheckState *healthcheck.CheckState
		}
		// CheckerAPIExt holds details about calls to the CheckerAPIExt method.
		CheckerAPIExt []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
		// GetDimensionOptions holds details about calls to the GetDimensionOptions method.
		GetDimensionOptions []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Req is the req argument value.
			Req cantabular.StaticDatasetQueryRequest
		}
	}
	lockChecker             sync.RWMutex
	lockCheckerAPIExt       sync.RWMutex
	lockGetDimensionOptions sync.RWMutex
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

// CheckerAPIExt calls CheckerAPIExtFunc.
func (mock *CantabularClientMock) CheckerAPIExt(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerAPIExtFunc == nil {
		panic("CantabularClientMock.CheckerAPIExtFunc: method is nil but CantabularClient.CheckerAPIExt was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}{
		Ctx:   ctx,
		State: state,
	}
	mock.lockCheckerAPIExt.Lock()
	mock.calls.CheckerAPIExt = append(mock.calls.CheckerAPIExt, callInfo)
	mock.lockCheckerAPIExt.Unlock()
	return mock.CheckerAPIExtFunc(ctx, state)
}

// CheckerAPIExtCalls gets all the calls that were made to CheckerAPIExt.
// Check the length with:
//     len(mockedCantabularClient.CheckerAPIExtCalls())
func (mock *CantabularClientMock) CheckerAPIExtCalls() []struct {
	Ctx   context.Context
	State *healthcheck.CheckState
} {
	var calls []struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}
	mock.lockCheckerAPIExt.RLock()
	calls = mock.calls.CheckerAPIExt
	mock.lockCheckerAPIExt.RUnlock()
	return calls
}

// GetDimensionOptions calls GetDimensionOptionsFunc.
func (mock *CantabularClientMock) GetDimensionOptions(ctx context.Context, req cantabular.StaticDatasetQueryRequest) (*cantabular.GetDimensionOptionsResponse, error) {
	if mock.GetDimensionOptionsFunc == nil {
		panic("CantabularClientMock.GetDimensionOptionsFunc: method is nil but CantabularClient.GetDimensionOptions was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Req cantabular.StaticDatasetQueryRequest
	}{
		Ctx: ctx,
		Req: req,
	}
	mock.lockGetDimensionOptions.Lock()
	mock.calls.GetDimensionOptions = append(mock.calls.GetDimensionOptions, callInfo)
	mock.lockGetDimensionOptions.Unlock()
	return mock.GetDimensionOptionsFunc(ctx, req)
}

// GetDimensionOptionsCalls gets all the calls that were made to GetDimensionOptions.
// Check the length with:
//     len(mockedCantabularClient.GetDimensionOptionsCalls())
func (mock *CantabularClientMock) GetDimensionOptionsCalls() []struct {
	Ctx context.Context
	Req cantabular.StaticDatasetQueryRequest
} {
	var calls []struct {
		Ctx context.Context
		Req cantabular.StaticDatasetQueryRequest
	}
	mock.lockGetDimensionOptions.RLock()
	calls = mock.calls.GetDimensionOptions
	mock.lockGetDimensionOptions.RUnlock()
	return calls
}
