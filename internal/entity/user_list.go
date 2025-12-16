package entity

import "codebase-app/pkg/types"

type GetUsersReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`

	types.MetaQuery
}

type GetUsersResp struct {
	Items []UserItem `json:"items"`
	Meta  types.Meta `json:"meta"`
}

type UserItem struct {
	ID     string `json:"id" db:"id"`
	RoleID string `json:"role_id" db:"role_id"`
	Name   string `json:"name" db:"name"`
	Email  string `json:"email" db:"email"`
	Role   string `json:"role" db:"role"`
}
