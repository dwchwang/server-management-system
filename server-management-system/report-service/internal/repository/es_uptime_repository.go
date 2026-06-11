package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/vcs-sms/report-service/internal/dto"
)

// UptimeCalculator defines the interface for computing uptime from Elasticsearch.
type UptimeCalculator interface {
	// GetUptimeSummary calculates average uptime for the given date range.
	GetUptimeSummary(ctx context.Context, startDate, endDate time.Time) (*dto.ReportSummaryResponse, error)

	// GetLowUptimeServers returns the top N servers with lowest uptime.
	GetLowUptimeServers(ctx context.Context, startDate, endDate time.Time, topN int) ([]dto.ServerUptime, error)
}

type esUptimeRepo struct {
	client *elasticsearch.Client
	index  string
}

// NewESUptimeRepo creates a new Elasticsearch uptime repository.
func NewESUptimeRepo(client *elasticsearch.Client, index string) UptimeCalculator {
	return &esUptimeRepo{client: client, index: index}
}

// GetUptimeSummary calculates average uptime for the given date range using ES aggregations.
func (r *esUptimeRepo) GetUptimeSummary(ctx context.Context, startDate, endDate time.Time) (*dto.ReportSummaryResponse, error) {
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{
						"range": map[string]interface{}{
							"checked_at": map[string]interface{}{
								"gte": startDate.Format(time.RFC3339),
								"lt":  endDate.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"per_server": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "server_id",
					"size":  10000,
				},
				"aggs": map[string]interface{}{
					"total_checks": map[string]interface{}{
						"value_count": map[string]interface{}{
							"field": "status",
						},
					},
					"on_checks": map[string]interface{}{
						"filter": map[string]interface{}{
							"term": map[string]interface{}{
								"status": "on",
							},
						},
					},
					"latest_check": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"size": 1,
							"sort": []interface{}{
								map[string]interface{}{
									"checked_at": map[string]interface{}{
										"order": "desc",
									},
								},
							},
							"_source": map[string]interface{}{
								"includes": []string{"status", "server_name"},
							},
						},
					},
					"uptime_rate": map[string]interface{}{
						"bucket_script": map[string]interface{}{
							"buckets_path": map[string]interface{}{
								"on":    "on_checks._count",
								"total": "total_checks",
							},
							"script": "params.on / params.total * 100",
						},
					},
				},
			},
			"avg_uptime": map[string]interface{}{
				"avg_bucket": map[string]interface{}{
					"buckets_path": "per_server>uptime_rate",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode ES query: %w", err)
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.index),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("ES search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES search returned error: %s", res.String())
	}

	return r.parseSummaryBody(res.Body, startDate, endDate)
}

// parseSummaryBody parses the ES aggregation response body into a ReportSummaryResponse.
func (r *esUptimeRepo) parseSummaryBody(body io.Reader, startDate, endDate time.Time) (*dto.ReportSummaryResponse, error) {
	var result struct {
		Aggregations struct {
			PerServer struct {
				Buckets []struct {
					Key         string `json:"key"`
					DocCount    int64  `json:"doc_count"`
					TotalChecks struct {
						Value float64 `json:"value"`
					} `json:"total_checks"`
					OnChecks struct {
						DocCount int64 `json:"doc_count"`
					} `json:"on_checks"`
					LatestCheck struct {
						Hits struct {
							Hits []struct {
								Source struct {
									Status     string `json:"status"`
									ServerName string `json:"server_name"`
								} `json:"_source"`
							} `json:"hits"`
						} `json:"hits"`
					} `json:"latest_check"`
					UptimeRate struct {
						Value float64 `json:"value"`
					} `json:"uptime_rate"`
				} `json:"buckets"`
			} `json:"per_server"`
			AvgUptime struct {
				Value float64 `json:"value"`
			} `json:"avg_uptime"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse ES response: %w", err)
	}

	buckets := result.Aggregations.PerServer.Buckets
	serversOn := 0
	serversOff := 0
	var totalChecks int64

	for _, b := range buckets {
		totalChecks += int64(b.TotalChecks.Value)
		// Determine current status from latest check
		if len(b.LatestCheck.Hits.Hits) > 0 {
			if b.LatestCheck.Hits.Hits[0].Source.Status == "on" {
				serversOn++
			} else {
				serversOff++
			}
		}
	}

	reportEndDate := endDate.AddDate(0, 0, -1)

	return &dto.ReportSummaryResponse{
		StartDate:    startDate.Format("2006-01-02"),
		EndDate:      reportEndDate.Format("2006-01-02"),
		TotalServers: len(buckets),
		ServersOn:    serversOn,
		ServersOff:   serversOff,
		AvgUptimePct: result.Aggregations.AvgUptime.Value,
		TotalChecks:  totalChecks,
	}, nil
}

// GetLowUptimeServers returns the top N servers with lowest uptime.
func (r *esUptimeRepo) GetLowUptimeServers(ctx context.Context, startDate, endDate time.Time, topN int) ([]dto.ServerUptime, error) {
	// Use same aggregation but collect all server uptimes
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{
						"range": map[string]interface{}{
							"checked_at": map[string]interface{}{
								"gte": startDate.Format(time.RFC3339),
								"lt":  endDate.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"per_server": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "server_id",
					"size":  10000,
				},
				"aggs": map[string]interface{}{
					"total_checks": map[string]interface{}{
						"value_count": map[string]interface{}{
							"field": "status",
						},
					},
					"on_checks": map[string]interface{}{
						"filter": map[string]interface{}{
							"term": map[string]interface{}{
								"status": "on",
							},
						},
					},
					"server_name": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"size": 1,
							"_source": map[string]interface{}{
								"includes": []string{"server_name"},
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode ES query: %w", err)
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.index),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("ES search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES search returned error: %s", res.String())
	}

	return r.parseLowUptimeBody(res.Body, topN)
}

// parseLowUptimeBody parses the ES response body and returns top N lowest uptime servers.
func (r *esUptimeRepo) parseLowUptimeBody(body io.Reader, topN int) ([]dto.ServerUptime, error) {
	var result struct {
		Aggregations struct {
			PerServer struct {
				Buckets []struct {
					Key         string `json:"key"`
					TotalChecks struct {
						Value float64 `json:"value"`
					} `json:"total_checks"`
					OnChecks struct {
						DocCount int64 `json:"doc_count"`
					} `json:"on_checks"`
					ServerName struct {
						Hits struct {
							Hits []struct {
								Source struct {
									ServerName string `json:"server_name"`
								} `json:"_source"`
							} `json:"hits"`
						} `json:"hits"`
					} `json:"server_name"`
				} `json:"buckets"`
			} `json:"per_server"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse ES response: %w", err)
	}

	var servers []dto.ServerUptime
	for _, b := range result.Aggregations.PerServer.Buckets {
		serverName := ""
		if len(b.ServerName.Hits.Hits) > 0 {
			serverName = b.ServerName.Hits.Hits[0].Source.ServerName
		}

		totalChecks := int64(b.TotalChecks.Value)
		uptimePct := 0.0
		if totalChecks > 0 {
			uptimePct = float64(b.OnChecks.DocCount) / float64(totalChecks) * 100.0
		}

		servers = append(servers, dto.ServerUptime{
			ServerID:    b.Key,
			ServerName:  serverName,
			UptimePct:   uptimePct,
			TotalChecks: totalChecks,
			OnChecks:    b.OnChecks.DocCount,
		})
	}

	// Sort by uptime ascending
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].UptimePct < servers[j].UptimePct
	})

	// Return top N
	if len(servers) > topN {
		servers = servers[:topN]
	}

	return servers, nil
}
