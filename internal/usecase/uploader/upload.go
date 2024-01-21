package uploader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"karma_s3/pkg/logger"
)

type Service struct {
	log             *logger.Logger
	dataService     DataSrv
	metadataManager MetadataManager
}

func NewService(
	logger *logger.Logger,
	dataSrv DataSrv,
	metadataManager MetadataManager,
) *Service {
	return &Service{
		log:             logger,
		dataService:     dataSrv,
		metadataManager: metadataManager,
	}
}

var (
	ErrObjNotFound      = errors.New("object not found")
	ErrObjAlreadyExists = errors.New("object already exits")
)

func (s *Service) UploadObject(ctx context.Context, obj []byte) (string, error) {
	objHash := s.makeHash(obj)
	found, err := s.metadataManager.IsExist(ctx, objHash)
	if err != nil {
		return "", fmt.Errorf("get metadata: %w", err)
	}
	if found {
		return "", ErrObjAlreadyExists
	}
	metadata, err := s.dataService.UploadObject(ctx, obj, objHash)
	if err != nil {
		return "", fmt.Errorf("upload object: %s, err: %w", objHash, err)
	}
	if err = s.metadataManager.SetMetadata(ctx, objHash, metadata); err != nil {
		// TODO send to daemon for retries
		return "", fmt.Errorf("set metadata: obj: %s, err: %w", objHash, err)
	}
	return metadata.ObjectName, nil
}

func (s *Service) makeHash(obj []byte) string {
	hasher := sha256.New()
	hasher.Write(obj)
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}
