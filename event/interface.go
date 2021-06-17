package event

import (
	"context"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(context.Context, *HelloCalled) error
}

type coder interface {
	Code() int
}

type dataLogger interface {
	LogData() map[string]interface{}
}