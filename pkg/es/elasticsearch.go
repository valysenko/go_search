package es

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

type ESConfig struct {
	Addresses []string
	Index     string
}

type Client struct {
	es    *elasticsearch.TypedClient
	index string
}

func NewClient(cfg *ESConfig) (*Client, error) {
	esConfig := elasticsearch.Config{
		Addresses: cfg.Addresses,
	}

	client, err := elasticsearch.NewTypedClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	return &Client{
		es:    client,
		index: cfg.Index,
	}, nil
}

func (c *Client) GetTypedClient() *elasticsearch.TypedClient {
	return c.es
}

func (c *Client) GetIndex() string {
	return c.index
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.es.Info().Do(ctx)
	return err
}

func (c *Client) Close(ctx context.Context) error {
	return c.es.Close(ctx)
}
