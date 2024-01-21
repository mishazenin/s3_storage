package metadata

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
	"karma_s3/internal/entity"
	metadataManager "karma_s3/internal/usecase/metadata_service"
)

type Client struct {
	client *redis.Client
	store  map[string]entity.Metadata
}

func NewClient() (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to metadata: %w", err)
	}
	return &Client{client: client, store: make(map[string]entity.Metadata)}, nil
}

func (c *Client) GetData(ctx context.Context, objHash string) (entity.Metadata, error) {
	val, ok := c.store[objHash]
	if !ok {
		return entity.Metadata{}, metadataManager.ErNotFound
	}
	return val, nil
}

func (c *Client) SetData(ctx context.Context, objHash string, metadata entity.Metadata) error {
	c.store[objHash] = metadata
	return nil
}
