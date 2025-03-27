package entity

type GetMeReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`
}

type GetMeResp struct {
	Id          string       `json:"id" db:"id"`
	RoleId      string       `json:"role_id" db:"role_id"`
	Role        string       `json:"role" db:"role"`
	Name        string       `json:"name" db:"name"`
	Email       string       `json:"email" db:"email"`
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}
