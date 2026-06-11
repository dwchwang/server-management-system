package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/vcs-sms/monitor-service/internal/checker"
)

// ESStatusLogRepo defines the interface for storing health-check results in Elasticsearch.
type ESStatusLogRepo interface {
	BulkIndex(ctx context.Context, results []*checker.HealthResult) error
}

type esStatusLogRepo struct {
	client *elasticsearch.Client
	index  string
}

// NewESStatusLogRepo creates a new Elasticsearch status log repository.
func NewESStatusLogRepo(client *elasticsearch.Client, index string) ESStatusLogRepo {
	return &esStatusLogRepo{client: client, index: index}
}

// BulkIndex writes multiple health-check results to Elasticsearch using the Bulk API.
func (r *esStatusLogRepo) BulkIndex(ctx context.Context, results []*checker.HealthResult) error {
	if len(results) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, result := range results {
		// Action metadata line
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": r.index,
			},
		}
		metaJSON, _ := json.Marshal(meta)
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		// Document line
		doc := map[string]interface{}{
			"server_id":        result.ServerID,
			"server_name":      result.ServerName,
			"status":           result.Status,
			"checked_at":       result.CheckedAt.Format(time.RFC3339),
			"response_time_ms": result.ResponseTimeMs,
			"check_method":     result.CheckMethod,
		}
		if result.Error != "" {
			doc["error"] = result.Error
		}
		docJSON, _ := json.Marshal(doc)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	// Execute bulk request
	res, err := r.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		r.client.Bulk.WithContext(ctx),
		r.client.Bulk.WithIndex(r.index),
	)
	if err != nil {
		return fmt.Errorf("bulk index request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index returned error: %s", res.String())
	}

	return nil
}
