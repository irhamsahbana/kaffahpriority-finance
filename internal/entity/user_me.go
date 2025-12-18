package entity

type GetMeReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`
}

type GetMeResp struct {
	ID          string       `json:"id" db:"id"`
	RoleID      string       `json:"role_id" db:"role_id"`
	Role        string       `json:"role" db:"role"`
	Name        string       `json:"name" db:"name"`
	Email       string       `json:"email" db:"email"`
	Permissions []Permission `json:"permissions"`
}
