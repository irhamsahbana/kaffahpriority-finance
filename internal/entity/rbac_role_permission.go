package entity

type UpdateRolePermissionsReq struct {
	UserID      string   `json:"user_id" validate:"required,ulid"`
	RoleID      string   `json:"role_id" validate:"required,ulid"`
	Permissions []string `json:"permissions" validate:"required,dive,ulid"`
}

type UpdateRolePermissionsResp struct {
	ID string `json:"id"`
}
