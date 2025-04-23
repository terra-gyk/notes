package repository

import (
  "gorm.io/gorm"
  "yourproject/internal/score_service/entity"
  "go.uber.org/zap"
)

// User 定义数据库中的用户表结构
type User struct {
  gorm.Model
  ID  string `gorm:"unique;not null"`
  Score int
}

// UserRepository 定义用户仓库接口
type UserRepository interface {
  GetUserScore(id string) (*entity.User, error)
  UpdateUserScore(id string, score int) error
}

// GormUserRepository 是用户仓库的 GORM 实现
type GormUserRepository struct {
  db   *gorm.DB
  logger *zap.Logger
}

// NewGormUserRepository 创建一个新的 GORM 用户仓库实例
func NewGormUserRepository(db *gorm.DB, logger *zap.Logger) *GormUserRepository {
  return &GormUserRepository{
    db:   db,
    logger: logger,
  }
}

// GetUserScore 根据用户 ID 获取用户分数
func (r *GormUserRepository) GetUserScore(id string) (*entity.User, error) {
  var user User
  result := r.db.Where("id =?", id).First(&user)
  if result.Error != nil {
    if result.Error == gorm.ErrRecordNotFound {
      return nil, nil
    }
    r.logger.Error("failed to get user score", zap.String("id", id), zap.Error(result.Error))
    return nil, result.Error
  }
  return &entity.User{
    ID:  user.ID,
    Score: user.Score,
  }, nil
}

// UpdateUserScore 更新用户分数
func (r *GormUserRepository) UpdateUserScore(id string, score int) error {
  result := r.db.Model(&User{}).Where("id =?", id).Update("score", score)
  if result.Error != nil {
    r.logger.Error("failed to update user score", zap.String("id", id), zap.Error(result.Error))
  }
  return result.Error
}
  