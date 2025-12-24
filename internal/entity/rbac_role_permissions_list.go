package entity

type GetRoleAndPermissionsReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`
}

type GetRoleAndPermissionsResp struct {
	Items []RolePermissionItem `json:"items"`
}

type RolePermissionItem struct {
	RoleID      string       `json:"role_id" db:"role_id"`
	Role        string       `json:"role" db:"role"`
	Permissions []Permission `json:"permissions" db:"permissions"`
}

type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GetRoleDetailReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`
	RoleID string `json:"role_id" validate:"required,ulid"`
}

type GetRoleDetailResp struct {
	RolePermissionItem
}
