package entity

type UpdateUserReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	Id string `json:"id" validate:"required,ulid"`

	RoleId   string `json:"role_id" validate:"required,ulid"`
	Name     string `json:"name" validate:"required,max=255,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"omitempty,min=8"`
}

type UpdateUserResp struct {
	Id string `json:"id"`
}

type DeleteUserReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	Id string `json:"id" validate:"required,ulid"`
}
