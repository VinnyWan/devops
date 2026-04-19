package service

import (
	"devops-platform/internal/pkg/logger"

	"go.uber.org/zap"
)

const cloudSyncPageSize = 100

// maxPaginationPages safety limit: max 200 pages per sync
const maxPaginationPages = 200

// paginateSync generic paginated fetch. fetchPage callback receives offset, returns item count and error.
// On single-page failure, logs error and skips to next page.
func (s *CloudAccountService) paginateSync(fetchPage func(offset int64) (int, error)) error {
	var offset int64
	for i := 0; i < maxPaginationPages; i++ {
		count, err := fetchPage(offset)
		if err != nil {
			if logger.Log != nil {
				logger.Log.Error("云同步分页拉取失败，跳过该页", zap.Int64("offset", offset), zap.Error(err))
			}
			offset += cloudSyncPageSize
			continue
		}
		if count < cloudSyncPageSize {
			return nil
		}
		offset += cloudSyncPageSize
	}
	return nil
}
