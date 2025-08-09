package entity

import (
	"mime/multipart"

	"github.com/shopspring/decimal"
)

type ImportLecturersWagesReq struct {
	UserId string `json:"user_id" validate:"required,ulid"`
	File   multipart.File

	Registrations []ImportedLecturersWages `json:"registrations" validate:"required,dive"`
}

type ImportedLecturersWages struct {
	RegistrationId  string           `json:"registration_id"`
	ProgramMeetings int              `json:"program_meetings"`
	FL              *decimal.Decimal `json:"foreign_learning_fee"`
	NL              *decimal.Decimal `json:"night_learning_fee"`
	InitialFee      decimal.Decimal  `json:"initial_fee"`
	IsFullFee       bool             `json:"is_full_fee"`
	Notes           *string          `json:"notes"`
}
