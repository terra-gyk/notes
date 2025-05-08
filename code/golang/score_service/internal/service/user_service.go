package service

import (
  "yourproject/internal/score_service/entity"
  "yourproject/internal/score_service/repository"
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

// GetUserScore 根据用户 ID 获取用户分数
func (s *UserService) GetUserScore(id string) (*entity.User, error) {
  return s.userRepository.GetUserScore(id)
}

// UpdateUserScore 更新用户分数
func (s *UserService) UpdateUserScore(id string, score int) error {
  return s.userRepository.UpdateUserScore(id, score)
}
  