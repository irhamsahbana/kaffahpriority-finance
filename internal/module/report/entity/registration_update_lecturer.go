package entity

type UpdateRegistrationLecturerReq struct {
	UserID string `json:"user_id" validate:"ulid"`

	ID         string `params:"id" validate:"ulid"`
	LecturerId string `json:"lecturer_id" validate:"ulid"`
}

type UpdateRegistrationLecturerResp struct {
	ID string `json:"id"`
}
