package entity

type GetUserReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	Id string `json:"id" validate:"required,ulid"`
}

type GetUserResp struct {
	Id     string `json:"id" db:"id"`
	RoleId string `json:"role_id" db:"role_id"`
	Name   string `json:"name" db:"name"`
	Email  string `json:"email" db:"email"`
}
