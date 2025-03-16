package entity

type CreateRoleReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	Name string `json:"name" validate:"required,min=3,max=255"`
}

type CreateRoleResp struct {
	Id string `json:"id"`
}

type UpdateRoleReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	Id   string `json:"id" validate:"required,ulid"`
	Name string `json:"name" validate:"required,min=3,max=255"`
}

type UpdateRoleResp struct {
	Id string `json:"id"`
}

type DeleteRoleReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`

	Id string `json:"id" validate:"required,ulid"`
}
