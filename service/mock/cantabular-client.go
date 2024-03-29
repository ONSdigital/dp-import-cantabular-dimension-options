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
// 			CheckerAPIExtFunc: func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
// 				panic("mock out the CheckerAPIExt method")
// 			},
// 			GetAggregatedDimensionOptionsFunc: func(contextMoqParam context.Context, getAggregatedDimensionOptionsRequest cantabular.GetAggregatedDimensionOptionsRequest) (*cantabular.GetAggregatedDimensionOptionsResponse, error) {
// 				panic("mock out the GetAggregatedDimensionOptions method")
// 			},
// 			GetDimensionOptionsFunc: func(contextMoqParam context.Context, getDimensionOptionsRequest cantabular.GetDimensionOptionsRequest) (*cantabular.GetDimensionOptionsResponse, error) {
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
	CheckerAPIExtFunc func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error

	// GetAggregatedDimensionOptionsFunc mocks the GetAggregatedDimensionOptions method.
	GetAggregatedDimensionOptionsFunc func(contextMoqParam context.Context, getAggregatedDimensionOptionsRequest cantabular.GetAggregatedDimensionOptionsRequest) (*cantabular.GetAggregatedDimensionOptionsResponse, error)

	// GetDimensionOptionsFunc mocks the GetDimensionOptions method.
	GetDimensionOptionsFunc func(contextMoqParam context.Context, getDimensionOptionsRequest cantabular.GetDimensionOptionsRequest) (*cantabular.GetDimensionOptionsResponse, error)

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
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// CheckState is the checkState argument value.
			CheckState *healthcheck.CheckState
		}
		// GetAggregatedDimensionOptions holds details about calls to the GetAggregatedDimensionOptions method.
		GetAggregatedDimensionOptions []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// GetAggregatedDimensionOptionsRequest is the getAggregatedDimensionOptionsRequest argument value.
			GetAggregatedDimensionOptionsRequest cantabular.GetAggregatedDimensionOptionsRequest
		}
		// GetDimensionOptions holds details about calls to the GetDimensionOptions method.
		GetDimensionOptions []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// GetDimensionOptionsRequest is the getDimensionOptionsRequest argument value.
			GetDimensionOptionsRequest cantabular.GetDimensionOptionsRequest
		}
	}
	lockChecker                       sync.RWMutex
	lockCheckerAPIExt                 sync.RWMutex
	lockGetAggregatedDimensionOptions sync.RWMutex
	lockGetDimensionOptions           sync.RWMutex
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
func (mock *CantabularClientMock) CheckerAPIExt(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
	if mock.CheckerAPIExtFunc == nil {
		panic("CantabularClientMock.CheckerAPIExtFunc: method is nil but CantabularClient.CheckerAPIExt was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}{
		ContextMoqParam: contextMoqParam,
		CheckState:      checkState,
	}
	mock.lockCheckerAPIExt.Lock()
	mock.calls.CheckerAPIExt = append(mock.calls.CheckerAPIExt, callInfo)
	mock.lockCheckerAPIExt.Unlock()
	return mock.CheckerAPIExtFunc(contextMoqParam, checkState)
}

// CheckerAPIExtCalls gets all the calls that were made to CheckerAPIExt.
// Check the length with:
//     len(mockedCantabularClient.CheckerAPIExtCalls())
func (mock *CantabularClientMock) CheckerAPIExtCalls() []struct {
	ContextMoqParam context.Context
	CheckState      *healthcheck.CheckState
} {
	var calls []struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}
	mock.lockCheckerAPIExt.RLock()
	calls = mock.calls.CheckerAPIExt
	mock.lockCheckerAPIExt.RUnlock()
	return calls
}

// GetAggregatedDimensionOptions calls GetAggregatedDimensionOptionsFunc.
func (mock *CantabularClientMock) GetAggregatedDimensionOptions(contextMoqParam context.Context, getAggregatedDimensionOptionsRequest cantabular.GetAggregatedDimensionOptionsRequest) (*cantabular.GetAggregatedDimensionOptionsResponse, error) {
	if mock.GetAggregatedDimensionOptionsFunc == nil {
		panic("CantabularClientMock.GetAggregatedDimensionOptionsFunc: method is nil but CantabularClient.GetAggregatedDimensionOptions was just called")
	}
	callInfo := struct {
		ContextMoqParam                      context.Context
		GetAggregatedDimensionOptionsRequest cantabular.GetAggregatedDimensionOptionsRequest
	}{
		ContextMoqParam:                      contextMoqParam,
		GetAggregatedDimensionOptionsRequest: getAggregatedDimensionOptionsRequest,
	}
	mock.lockGetAggregatedDimensionOptions.Lock()
	mock.calls.GetAggregatedDimensionOptions = append(mock.calls.GetAggregatedDimensionOptions, callInfo)
	mock.lockGetAggregatedDimensionOptions.Unlock()
	return mock.GetAggregatedDimensionOptionsFunc(contextMoqParam, getAggregatedDimensionOptionsRequest)
}

// GetAggregatedDimensionOptionsCalls gets all the calls that were made to GetAggregatedDimensionOptions.
// Check the length with:
//     len(mockedCantabularClient.GetAggregatedDimensionOptionsCalls())
func (mock *CantabularClientMock) GetAggregatedDimensionOptionsCalls() []struct {
	ContextMoqParam                      context.Context
	GetAggregatedDimensionOptionsRequest cantabular.GetAggregatedDimensionOptionsRequest
} {
	var calls []struct {
		ContextMoqParam                      context.Context
		GetAggregatedDimensionOptionsRequest cantabular.GetAggregatedDimensionOptionsRequest
	}
	mock.lockGetAggregatedDimensionOptions.RLock()
	calls = mock.calls.GetAggregatedDimensionOptions
	mock.lockGetAggregatedDimensionOptions.RUnlock()
	return calls
}

// GetDimensionOptions calls GetDimensionOptionsFunc.
func (mock *CantabularClientMock) GetDimensionOptions(contextMoqParam context.Context, getDimensionOptionsRequest cantabular.GetDimensionOptionsRequest) (*cantabular.GetDimensionOptionsResponse, error) {
	if mock.GetDimensionOptionsFunc == nil {
		panic("CantabularClientMock.GetDimensionOptionsFunc: method is nil but CantabularClient.GetDimensionOptions was just called")
	}
	callInfo := struct {
		ContextMoqParam            context.Context
		GetDimensionOptionsRequest cantabular.GetDimensionOptionsRequest
	}{
		ContextMoqParam:            contextMoqParam,
		GetDimensionOptionsRequest: getDimensionOptionsRequest,
	}
	mock.lockGetDimensionOptions.Lock()
	mock.calls.GetDimensionOptions = append(mock.calls.GetDimensionOptions, callInfo)
	mock.lockGetDimensionOptions.Unlock()
	return mock.GetDimensionOptionsFunc(contextMoqParam, getDimensionOptionsRequest)
}

// GetDimensionOptionsCalls gets all the calls that were made to GetDimensionOptions.
// Check the length with:
//     len(mockedCantabularClient.GetDimensionOptionsCalls())
func (mock *CantabularClientMock) GetDimensionOptionsCalls() []struct {
	ContextMoqParam            context.Context
	GetDimensionOptionsRequest cantabular.GetDimensionOptionsRequest
} {
	var calls []struct {
		ContextMoqParam            context.Context
		GetDimensionOptionsRequest cantabular.GetDimensionOptionsRequest
	}
	mock.lockGetDimensionOptions.RLock()
	calls = mock.calls.GetDimensionOptions
	mock.lockGetDimensionOptions.RUnlock()
	return calls
}
