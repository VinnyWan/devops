package repository

import (
	"time"

	"devops-platform/internal/modules/log/model"
)

type LogRepo struct{}

func NewLogRepo() *LogRepo {
	return &LogRepo{}
}

func (r *LogRepo) List() []model.LogEntry {
	now := time.Now()
	return []model.LogEntry{
		{
			ID:        1,
			Cluster:   "default",
			Namespace: "payments",
			Pod:       "payments-api-6f4f7d47f9-abcde",
			Source:    "stdout",
			Level:     "info",
			Message:   "request completed status=200 path=/api/v1/payments",
			CreatedAt: now.Add(-5 * time.Minute),
		},
		{
			ID:        2,
			Cluster:   "default",
			Namespace: "payments",
			Pod:       "payments-worker-7cf8d9555-jk2ls",
			Source:    "stderr",
			Level:     "error",
			Message:   "failed to consume message: timeout",
			CreatedAt: now.Add(-4 * time.Minute),
		},
		{
			ID:        3,
			Cluster:   "prod-sh",
			Namespace: "gateway",
			Pod:       "gateway-5f6459676f-cv9hm",
			Source:    "stdout",
			Level:     "warn",
			Message:   "upstream latency exceeded threshold",
			CreatedAt: now.Add(-2 * time.Minute),
		},
	}
}
