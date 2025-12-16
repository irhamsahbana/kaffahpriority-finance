package entity

type UpdateUserReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`

	ID string `json:"id" validate:"required,ulid"`

	RoleID   string `json:"role_id" validate:"required,ulid"`
	Name     string `json:"name" validate:"required,max=255,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"omitempty,min=8"`
}

type UpdateUserResp struct {
	ID string `json:"id"`
}
