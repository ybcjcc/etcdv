package etcdhistory

import (
	"context"
	"testing"
	"time"

	"github.com/ybcjcc/etcdv/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestGetVersionHistory(t *testing.T) {
	// 设置测试用例
	tests := []struct {
		name    string
		opts    types.Options
		wantErr bool
	}{
		{
			name: "基本查询测试",
			opts: types.Options{
				Endpoints: []string{"localhost:2379"},
				Key:       "test-key",
				Order:     "desc",
			},
			wantErr: false,
		},
		{
			name: "带限制的查询测试",
			opts: types.Options{
				Endpoints: []string{"localhost:2379"},
				Key:       "test-key",
				Limit:     5,
				Order:     "asc",
			},
			wantErr: false,
		},
	}

	// 运行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试上下文
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 执行测试
			records, err := GetVersionHistory(ctx, tt.opts)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, records)

			// 验证记录数量
			if tt.opts.Limit > 0 {
				assert.LessOrEqual(t, len(records), tt.opts.Limit)
			}

			// 验证排序
			if len(records) > 1 {
				for i := 0; i < len(records)-1; i++ {
					if tt.opts.Order == "asc" {
						assert.Less(t, records[i].Revision, records[i+1].Revision)
					} else {
						assert.Greater(t, records[i].Revision, records[i+1].Revision)
					}
				}
			}
		})
	}
}

func BenchmarkGetVersionHistory(b *testing.B) {
	// 设置基准测试选项
	opts := types.Options{
		Endpoints: []string{"localhost:2379"},
		Key:       "bench-key",
		Order:     "desc",
	}

	// 运行基准测试
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := GetVersionHistory(context.Background(), opts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
