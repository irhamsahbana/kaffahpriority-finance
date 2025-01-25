package entity

type CopyRegistrationsReq struct {
	UserId        string          `json:"user_id" validate:"required,ulid"`
	Registrations []CopyRegisItem `validate:"required,dive"`
}

type CopyRegisItem struct {
	RegisId string `json:"registration_id" validate:"required,ulid"`
}
