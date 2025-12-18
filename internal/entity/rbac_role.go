package entity

type CreateRoleReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`

	Name        string   `json:"name" validate:"required,min=3,max=255"`
	Permissions []string `json:"permissions" validate:"required,dive,ulid"`
}

type CreateRoleResp struct {
	ID string `json:"id"`
}

type UpdateRoleReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`

	ID          string   `json:"id" validate:"required,ulid"`
	Name        string   `json:"name" validate:"required,min=3,max=255"`
	Permissions []string `json:"permissions" validate:"required,dive,ulid"`
}

type UpdateRoleResp struct {
	ID string `json:"id"`
}

type DeleteRoleReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`

	ID string `json:"id" validate:"required,ulid"`
}
