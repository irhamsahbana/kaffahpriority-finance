package ports

import (
	"codebase-app/internal/entity"
	"context"
)

type UserRepository interface {
	Login(ctx context.Context, req *entity.LoginReq) (*entity.LoginResp, error)
	GetMe(ctx context.Context, req *entity.GetMeReq) (*entity.GetMeResp, error)

	GetUsers(ctx context.Context, req *entity.GetUsersReq) (*entity.GetUsersResp, error)
	GetUser(ctx context.Context, req *entity.GetUserReq) (*entity.GetUserResp, error)
	CreateUser(ctx context.Context, req *entity.CreateUserReq) (*entity.CreateUserResp, error)
	UpdateUser(ctx context.Context, req *entity.UpdateUserReq) (*entity.UpdateUserResp, error)
	DeleteUser(ctx context.Context, req *entity.DeleteUserReq) error
}

type UserService interface {
	Login(ctx context.Context, req *entity.LoginReq) (*entity.LoginResp, error)
	GetMe(ctx context.Context, req *entity.GetMeReq) (*entity.GetMeResp, error)

	GetUsers(ctx context.Context, req *entity.GetUsersReq) (*entity.GetUsersResp, error)
	GetUser(ctx context.Context, req *entity.GetUserReq) (*entity.GetUserResp, error)
	CreateUser(ctx context.Context, req *entity.CreateUserReq) (*entity.CreateUserResp, error)
	UpdateUser(ctx context.Context, req *entity.UpdateUserReq) (*entity.UpdateUserResp, error)
	DeleteUser(ctx context.Context, req *entity.DeleteUserReq) error
}
