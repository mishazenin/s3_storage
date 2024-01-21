package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"

	"karma_s3/pkg/logger"
)

type Service struct {
	log             *logger.Logger
	dataService     DataService
	metadataManager MetadataManager
}

func NewService(
	logger *logger.Logger,
	dataService DataService,
	metadataManager MetadataManager,
) *Service {
	return &Service{
		log:             logger,
		dataService:     dataService,
		metadataManager: metadataManager,
	}
}

var ErrObjectNotExist = errors.New("object does not exists")

func (s *Service) DownloadObject(ctx context.Context, objName string) (io.Reader, error) {
	metadata, err := s.metadataManager.GetMetadata(ctx, objName)
	if err != nil {
		return nil, fmt.Errorf("get metadata: %w", err)
	}
	obj, err := s.dataService.GetObject(ctx, metadata)
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	return obj, nil
}
