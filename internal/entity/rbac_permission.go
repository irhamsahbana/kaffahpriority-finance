package entity

type GetPermissionsReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`
}

type GetPermissionsResp struct {
	Items []Permission `json:"items"`
}
