package entity

type UpdateRolePermissionsReq struct {
	UserId      string   `json:"user_id" validate:"required,ulid"`
	RoleId      string   `json:"role_id" validate:"required,ulid"`
	Permissions []string `json:"permissions" validate:"required,dive,ulid"`
}

type UpdateRolePermissionsResp struct {
	Id string `json:"id"`
}
