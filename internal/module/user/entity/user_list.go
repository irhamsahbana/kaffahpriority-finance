package entity

import "codebase-app/pkg/types"

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
