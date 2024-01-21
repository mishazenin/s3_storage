package downloader

import (
	"context"
	"io"

	"karma_s3/internal/entity"
)

type (
	DataService interface {
		GetObject(ctx context.Context, metadata entity.Metadata) (io.Reader, error)
	}
	MetadataManager interface {
		GetMetadata(ctx context.Context, objHash string) (entity.Metadata, error)
	}
)
