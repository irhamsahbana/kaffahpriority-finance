package entity

type CreateRegistrationsReq struct {
	UserId        string             `json:"user_id" validate:"required,ulid"`
	Registrations []RegistrationItem `validate:"required,dive"`
}

type RegistrationItem struct {
	TemplateId          string `json:"template_id" validate:"required,ulid"`
	IsFirstRegistration bool   `json:"is_first_registration"`
}
