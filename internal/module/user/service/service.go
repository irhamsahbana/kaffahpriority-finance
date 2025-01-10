package service

import (
	"codebase-app/internal/module/user/entity"
	"codebase-app/internal/module/user/ports"
	"context"
)

var _ ports.UserService = &userService{}

type userService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) *userService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) Login(ctx context.Context, req *entity.LoginReq) (*entity.LoginResp, error) {
	return s.repo.Login(ctx, req)
}
