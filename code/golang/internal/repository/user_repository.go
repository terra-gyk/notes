package repository

import "yourproject/internal/entity"

// UserRepository 定义用户仓库接口
type UserRepository interface {
    GetUser(id string) (*entity.User, error)
    CreateUser(user *entity.User) error
}

// InMemoryUserRepository 是用户仓库的内存实现
type InMemoryUserRepository struct {
    users map[string]*entity.User
}

// NewInMemoryUserRepository 创建一个新的内存用户仓库实例
func NewInMemoryUserRepository() *InMemoryUserRepository {
    return &InMemoryUserRepository{
        users: make(map[string]*entity.User),
    }
}

// GetUser 根据用户 ID 获取用户信息
func (r *InMemoryUserRepository) GetUser(id string) (*entity.User, error) {
    user, exists := r.users[id]
    if!exists {
        return nil, nil
    }
    return user, nil
}

// CreateUser 创建一个新用户
func (r *InMemoryUserRepository) CreateUser(user *entity.User) error {
    r.users[user.ID] = user
    return nil
}
    