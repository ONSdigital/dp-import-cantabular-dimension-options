// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"sync"
)

var (
	lockCantabularClientMockGetCodebook sync.RWMutex
)

// CantabularClientMock is a mock implementation of handler.CantabularClient.
//
//     func TestSomethingThatUsesCantabularClient(t *testing.T) {
//
//         // make and configure a mocked handler.CantabularClient
//         mockedCantabularClient := &CantabularClientMock{
//             GetCodebookFunc: func(in1 context.Context, in2 cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
// 	               panic("mock out the GetCodebook method")
//             },
//         }
//
//         // use mockedCantabularClient in code that requires handler.CantabularClient
//         // and then make assertions.
//
//     }
type CantabularClientMock struct {
	// GetCodebookFunc mocks the GetCodebook method.
	GetCodebookFunc func(in1 context.Context, in2 cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetCodebook holds details about calls to the GetCodebook method.
		GetCodebook []struct {
			// In1 is the in1 argument value.
			In1 context.Context
			// In2 is the in2 argument value.
			In2 cantabular.GetCodebookRequest
		}
	}
}

// GetCodebook calls GetCodebookFunc.
func (mock *CantabularClientMock) GetCodebook(in1 context.Context, in2 cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
	if mock.GetCodebookFunc == nil {
		panic("CantabularClientMock.GetCodebookFunc: method is nil but CantabularClient.GetCodebook was just called")
	}
	callInfo := struct {
		In1 context.Context
		In2 cantabular.GetCodebookRequest
	}{
		In1: in1,
		In2: in2,
	}
	lockCantabularClientMockGetCodebook.Lock()
	mock.calls.GetCodebook = append(mock.calls.GetCodebook, callInfo)
	lockCantabularClientMockGetCodebook.Unlock()
	return mock.GetCodebookFunc(in1, in2)
}

// GetCodebookCalls gets all the calls that were made to GetCodebook.
// Check the length with:
//     len(mockedCantabularClient.GetCodebookCalls())
func (mock *CantabularClientMock) GetCodebookCalls() []struct {
	In1 context.Context
	In2 cantabular.GetCodebookRequest
} {
	var calls []struct {
		In1 context.Context
		In2 cantabular.GetCodebookRequest
	}
	lockCantabularClientMockGetCodebook.RLock()
	calls = mock.calls.GetCodebook
	lockCantabularClientMockGetCodebook.RUnlock()
	return calls
}
