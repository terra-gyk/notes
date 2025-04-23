package service

import (
  "yourproject/internal/name_service/entity"
  "yourproject/internal/name_service/repository"
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

// GetUserByName 根据用户 ID 获取用户姓名
func (s *UserService) GetUserByName(id string) (*entity.User, error) {
  return s.userRepository.GetUserByName(id)
}

// UpdateUserName 更新用户姓名
func (s *UserService) UpdateUserName(id, name string) error {
  return s.userRepository.UpdateUserName(id, name)
}
  