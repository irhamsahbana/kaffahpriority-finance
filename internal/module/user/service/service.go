package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	ports "codebase-app/internal/ports/module/user"
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
	ctx, span := tracing.StartSpan(ctx, "service.Login")
	defer span.End()
	return s.repo.Login(ctx, req)
}

func (s *userService) GetMe(ctx context.Context, req *entity.GetMeReq) (*entity.GetMeResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetMe")
	defer span.End()
	return s.repo.GetMe(ctx, req)
}

func (s *userService) GetUsers(ctx context.Context, req *entity.GetUsersReq) (*entity.GetUsersResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetUsers")
	defer span.End()
	return s.repo.GetUsers(ctx, req)
}

func (s *userService) GetUser(ctx context.Context, req *entity.GetUserReq) (*entity.GetUserResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetUser")
	defer span.End()
	return s.repo.GetUser(ctx, req)
}

func (s *userService) CreateUser(ctx context.Context, req *entity.CreateUserReq) (*entity.CreateUserResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreateUser")
	defer span.End()
	return s.repo.CreateUser(ctx, req)
}

func (s *userService) UpdateUser(ctx context.Context, req *entity.UpdateUserReq) (*entity.UpdateUserResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateUser")
	defer span.End()
	return s.repo.UpdateUser(ctx, req)
}

func (s *userService) DeleteUser(ctx context.Context, req *entity.DeleteUserReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeleteUser")
	defer span.End()
	return s.repo.DeleteUser(ctx, req)
}
