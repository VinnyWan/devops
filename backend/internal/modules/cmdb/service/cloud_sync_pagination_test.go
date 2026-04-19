package service

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaginateSync_SinglePage(t *testing.T) {
	svc := &CloudAccountService{}
	called := int32(0)
	err := svc.paginateSync(func(offset int64) (int, error) {
		atomic.AddInt32(&called, 1)
		return 50, nil
	})
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&called))
}

func TestPaginateSync_MultiplePages(t *testing.T) {
	svc := &CloudAccountService{}
	pages := []int{100, 100, 100, 30}
	called := int32(0)
	err := svc.paginateSync(func(offset int64) (int, error) {
		idx := int(atomic.LoadInt32(&called))
		atomic.AddInt32(&called, 1)
		return pages[idx], nil
	})
	require.NoError(t, err)
	require.Equal(t, int32(4), atomic.LoadInt32(&called))
}

func TestPaginateSync_PageErrorSkipped(t *testing.T) {
	svc := &CloudAccountService{}
	called := int32(0)
	err := svc.paginateSync(func(offset int64) (int, error) {
		idx := atomic.AddInt32(&called, 1)
		if idx == 1 {
			return 0, errors.New("API error")
		}
		return 50, nil
	})
	require.NoError(t, err)
	require.Equal(t, int32(2), atomic.LoadInt32(&called))
}

func TestPaginateSync_MaxPagesSafety(t *testing.T) {
	svc := &CloudAccountService{}
	called := int32(0)
	err := svc.paginateSync(func(offset int64) (int, error) {
		atomic.AddInt32(&called, 1)
		return 100, nil
	})
	require.NoError(t, err)
	require.Equal(t, int32(maxPaginationPages), atomic.LoadInt32(&called))
}
