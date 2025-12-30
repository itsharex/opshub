package biz

import (
	"context"

	"gorm.io/gorm"

	"github.com/ydcloud-dy/opshub/internal/data"
)

// UseCase 业务用例接口
type UseCase interface {
	// 在这里定义业务方法
}

// Biz 业务层
type Biz struct {
	db  *gorm.DB
	rdb *data.Redis
}

// NewBiz 创建业务层
func NewBiz(data *data.Data, redis *data.Redis) *Biz {
	return &Biz{
		db:  data.DB(),
		rdb: redis,
	}
}

// Example 示例方法
func (b *Biz) Example(ctx context.Context) error {
	// 业务逻辑实现
	return nil
}
