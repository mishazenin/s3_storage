package uploader

import (
	"context"

	"karma_s3/internal/entity"
)

type (
	DataSrv interface {
		UploadObject(ctx context.Context, obj []byte, objHash string) (entity.Metadata, error)
	}
	MetadataManager interface {
		SetMetadata(ctx context.Context, objHash string, metadata entity.Metadata) error
		IsExist(ctx context.Context, objHash string) (bool, error)
	}
)
