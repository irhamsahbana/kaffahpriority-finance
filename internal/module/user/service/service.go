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

func (s *userService) GetUsers(ctx context.Context, req *entity.GetUsersReq) (*entity.GetUsersResp, error) {
	return s.repo.GetUsers(ctx, req)
}

func (s *userService) GetUser(ctx context.Context, req *entity.GetUserReq) (*entity.GetUserResp, error) {
	return s.repo.GetUser(ctx, req)
}

func (s *userService) UpdateUser(ctx context.Context, req *entity.UpdateUserReq) (*entity.UpdateUserResp, error) {
	return s.repo.UpdateUser(ctx, req)
}

func (s *userService) DeleteUser(ctx context.Context, req *entity.DeleteUserReq) error {
	return s.repo.DeleteUser(ctx, req)
}
