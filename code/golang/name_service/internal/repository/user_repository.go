package repository

import (
  "gorm.io/gorm"
  "yourproject/internal/name_service/entity"
  "go.uber.org/zap"
)

// User 定义数据库中的用户表结构
type User struct {
  gorm.Model
  ID   string `gorm:"unique;not null"`
  Name string
}

// UserRepository 定义用户仓库接口
type UserRepository interface {
  GetUserByName(id string) (*entity.User, error)
  UpdateUserName(id, name string) error
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

// GetUserByName 根据用户 ID 获取用户姓名
func (r *GormUserRepository) GetUserByName(id string) (*entity.User, error) {
  var user User
  result := r.db.Where("id =?", id).First(&user)
  if result.Error != nil {
    if result.Error == gorm.ErrRecordNotFound {
      return nil, nil
    }
    r.logger.Error("failed to get user name", zap.String("id", id), zap.Error(result.Error))
    return nil, result.Error
  }
  return &entity.User{
    ID:   user.ID,
    Name: user.Name,
  }, nil
}

// UpdateUserName 更新用户姓名
func (r *GormUserRepository) UpdateUserName(id, name string) error {
  result := r.db.Model(&User{}).Where("id =?", id).Update("name", name)
  if result.Error != nil {
    r.logger.Error("failed to update user name", zap.String("id", id), zap.Error(result.Error))
  }
  return result.Error
}
  