package etcdhistory

import (
	"context"
	"fmt"
	"sort"
	"time"

	"etcdv/internal/utils"
	"etcdv/pkg/types"

	"go.etcd.io/etcd/client/v3"
)

// GetVersionHistory 获取指定键的历史版本记录
func GetVersionHistory(ctx context.Context, opts types.Options) ([]types.VersionRecord, error) {
	config := clientv3.Config{
		Endpoints:   opts.Endpoints,
		DialTimeout: 5 * time.Second,
	}

	// 如果提供了认证信息，则配置认证
	if opts.Username != "" {
		config.Username = opts.Username
		config.Password = opts.Password
	}

	cli, err := clientv3.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %v", err)
	}
	defer cli.Close()

	// 获取当前版本
	resp, err := cli.Get(ctx, opts.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %v", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("key '%s' not found", opts.Key)
	}

	// 获取历史版本
	var records []types.VersionRecord
	createRev := resp.Kvs[0].CreateRevision
	modRev := resp.Kvs[0].ModRevision

	// 根据order参数决定遍历方向
	start, end, step := modRev, createRev, -1
	if opts.Order == "asc" {
		start, end, step = createRev, modRev, 1
	}

	// 计算需要获取的版本数量
	revisionCount := (end - start) / int64(step)
	if revisionCount < 0 {
		revisionCount = -revisionCount
	}
	revisionCount++ // 包含起始版本

	// 如果设置了limit，调整版本数量
	if opts.Limit > 0 && int64(opts.Limit) < revisionCount {
		revisionCount = int64(opts.Limit)
	}

	// 创建版本号切片
	revisions := make([]int64, 0, revisionCount)
	for rev, count := start, int64(0); count < revisionCount; rev += int64(step) {
		revisions = append(revisions, rev)
		count++
	}

	// 使用批量查询获取历史版本
	resultChan := make(chan types.VersionRecord, len(revisions))
	errChan := make(chan error, len(revisions))
	semaphore := make(chan struct{}, 10) // 限制并发数

	// 并发获取历史版本
	for _, rev := range revisions {
		semaphore <- struct{}{}
		go func(revision int64) {
			defer func() { <-semaphore }()

			retryOpts := utils.DefaultRetryOptions()
			err := utils.RetryWithContext(ctx, retryOpts, func() error {
				getResp, err := cli.Get(ctx, opts.Key, clientv3.WithRev(revision))
				if err != nil {
					return err
				}

				if len(getResp.Kvs) > 0 {
					resultChan <- types.VersionRecord{
						Version:        getResp.Kvs[0].Version,
						Value:          string(getResp.Kvs[0].Value),
						Revision:       revision,
						CreateRevision: getResp.Kvs[0].CreateRevision,
						ModRevision:    getResp.Kvs[0].ModRevision,
					}
				}
				return nil
			})

			if err != nil {
				errChan <- fmt.Errorf("failed to get history at revision %d: %v", revision, err)
			}
		}(rev)
	}

	// 等待所有goroutine完成
	for range revisions {
		select {
		case record := <-resultChan:
			records = append(records, record)
		case err := <-errChan:
			return nil, err
		}
	}

	// 按照指定顺序排序结果
	sort.Slice(records, func(i, j int) bool {
		if opts.Order == "asc" {
			return records[i].Revision < records[j].Revision
		}
		return records[i].Revision > records[j].Revision
	})

	return records, nil
}
