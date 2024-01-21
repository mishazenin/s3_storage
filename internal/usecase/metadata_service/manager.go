package metadataManager

import (
	"context"
	"errors"

	"github.com/cenkalti/backoff"
	"karma_s3/internal/entity"
	"karma_s3/internal/usecase/uploader"
	"karma_s3/pkg/logger"
)

const mqxRetires = 2

type MetadataManager struct {
	log   *logger.Logger
	store Repository
}

func NewMetadataManager(
	log *logger.Logger,
	store Repository,
) *MetadataManager {
	return &MetadataManager{
		log:   log,
		store: store,
	}
}

var ErNotFound = errors.New("object not found")

func (m *MetadataManager) GetMetadata(ctx context.Context, objHash string) (entity.Metadata, error) {
	err := backoff.Retry(func() error {
		// TODO
		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), mqxRetires))
	if err != nil {
		return entity.Metadata{}, err
	}
	return entity.Metadata{}, uploader.ErrObjNotFound
}

func (m *MetadataManager) IsExist(ctx context.Context, objHash string) (bool, error) {
	return false, nil
}

func (m *MetadataManager) SetMetadata(ctx context.Context, objHash string, metadata entity.Metadata) error {
	return backoff.Retry(func() error {
		// TODO
		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), mqxRetires))
}
