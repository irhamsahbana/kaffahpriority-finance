package ports

import (
	"codebase-app/internal/module/user/entity"
	"context"
)

type UserRepository interface {
	Login(ctx context.Context, req *entity.LoginReq) (*entity.LoginResp, error)

	GetUsers(ctx context.Context, req *entity.GetUsersReq) (*entity.GetUsersResp, error)
}

type UserService interface {
	Login(ctx context.Context, req *entity.LoginReq) (*entity.LoginResp, error)

	GetUsers(ctx context.Context, req *entity.GetUsersReq) (*entity.GetUsersResp, error)
}
