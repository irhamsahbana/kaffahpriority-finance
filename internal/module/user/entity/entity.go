package entity

import "codebase-app/pkg/types"

type LoginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (r *LoginReq) Log() map[string]interface{} {
	return map[string]interface{}{
		"email": r.Email,
	}
}

type LoginResp struct {
	AccessToken string `json:"access_token"`
}

type GetUsersReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	types.MetaQuery
}

type GetUsersResp struct {
	Items []UserItem `json:"items"`
	Meta  types.Meta `json:"meta"`
}

type UserItem struct {
	Id     string `json:"id" db:"id"`
	RoleId string `json:"role_id" db:"role_id"`
	Name   string `json:"name" db:"name"`
	Role   string `json:"role" db:"role"`
}

type GetUserReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	Id string `json:"id" validate:"required,ulid"`
}

type GetUserResp struct {
	Id     string `json:"id" db:"id"`
	RoleId string `json:"role_id" db:"role_id"`
	Name   string `json:"name" db:"name"`
	Email  string `json:"email" db:"email"`
}

type UpdateUserReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	Id string `json:"id" validate:"required,ulid"`

	RoleId string `json:"role_id" validate:"required,ulid"`
	Name   string `json:"name" validate:"required,max=255,min=3"`
	Email  string `json:"email" validate:"required,email"`
}

type UpdateUserResp struct {
	Id string `json:"id"`
}

type DeleteUserReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	Id string `json:"id" validate:"required,ulid"`
}
