package entity

type GenerateRegistrationsReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`
}
