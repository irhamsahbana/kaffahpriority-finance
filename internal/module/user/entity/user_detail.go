package entity

type GetUserReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`

	ID string `json:"id" validate:"required,ulid"`
}

type GetUserResp struct {
	ID     string `json:"id" db:"id"`
	RoleID string `json:"role_id" db:"role_id"`
	Name   string `json:"name" db:"name"`
	Email  string `json:"email" db:"email"`
}
