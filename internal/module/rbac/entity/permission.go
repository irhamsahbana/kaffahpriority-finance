package entity

type GetPermissionsReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`
}

type GetPermissionsResp struct {
	Items []Permission `json:"items"`
}
