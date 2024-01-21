package metadataManager

import (
	"context"

	"karma_s3/internal/entity"
)

type Repository interface {
	GetData(ctx context.Context, objHash string) (entity.Metadata, error)
	SetData(ctx context.Context, objHash string, metadata entity.Metadata) error
}
