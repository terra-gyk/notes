package service

import (
    "yourproject/internal/entity"
    "yourproject/internal/repository"
)

// UserService 定义用户服务
type UserService struct {
    userRepository repository.UserRepository
}

// NewUserService 创建一个新的用户服务实例
func NewUserService(userRepository repository.UserRepository) *UserService {
    return &UserService{
        userRepository: userRepository,
    }
}

// GetUser 根据用户 ID 获取用户信息
func (s *UserService) GetUser(id string) (*entity.User, error) {
    return s.userRepository.GetUser(id)
}

// CreateUser 创建一个新用户
func (s *UserService) CreateUser(user *entity.User) error {
    return s.userRepository.CreateUser(user)
}
    