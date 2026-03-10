package api

import (
	"context"
	"log"

	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/google/uuid"
)

func (s *Server) SelectKey(ctx context.Context, req SelectKeyRequestObject) (SelectKeyResponseObject, error) {
	u, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return SelectKey401JSONResponse{Message: "Unauthorized"}, nil
	}
	source, err := s.store.GetSource(ctx, uuid.UUID(req.Body.SourceId))
	if err != nil {
		return SelectKey404JSONResponse{Message: "Source not found"}, nil
	}
	if source.UserID != u.ID {
		return SelectKey403JSONResponse{Message: "Forbidden: You do not own this source"}, nil
	}
	csvMeta, ok := source.Metadata.(*models.CSVSourceMetadata)
	if !ok {
		return SelectKey500JSONResponse{Message: "Source metadata not found"}, nil
	}
	return s.selectKeyForJob(ctx, csvMeta.FileURI, req.Body.ColumnsMetadata)
}

func (s *Server) selectKeyForJob(ctx context.Context, filePath string, meta *[]ColumnMetadata) (SelectKeyResponseObject, error) {
	csvResult, err := s.gcsReader.ReadCSV(ctx, filePath)
	if err != nil {
		log.Printf("Failed to read CSV file: %v", err)
		return SelectKey500JSONResponse{Message: "Failed to read CSV file"}, nil
	}
	if len(csvResult.Headers) == 0 {
		return SelectKey400JSONResponse{Message: "No headers found in CSV file"}, nil
	}
	result, err := s.keySelector.SelectBestKey(ctx, csvResult.Headers, toModelColumnMetadataSlicePtr(meta))
	if err != nil {
		log.Printf("Failed to select key: %v", err)
		return SelectKey500JSONResponse{Message: "Failed to select key"}, nil
	}
	return SelectKey200JSONResponse{SelectedKey: result.SelectedKey, AllKeys: result.AllKeys, Reasoning: result.Reasoning}, nil
}
