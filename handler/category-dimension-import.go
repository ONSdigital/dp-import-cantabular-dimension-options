package handler

import (
	"context"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/log.go/v2/log"
)

// CategoryDimensionImport is the handle for the CategoryDimensionImport event
type CategoryDimensionImport struct {
	cfg config.Config
}

// NewCategoryDimensionImport creates a new CategoryDimensionImport with the provided config
func NewCategoryDimensionImport(cfg config.Config) *CategoryDimensionImport {
	return &CategoryDimensionImport{
		cfg: cfg,
	}
}

// Handle takes a single event
func (h *CategoryDimensionImport) Handle(ctx context.Context, e *event.CategoryDimensionImport) error {
	logData := log.Data{
		"event": e,
	}
	log.Info(ctx, "event handler called", logData)

	// TODO process message

	return nil
}
