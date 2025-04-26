package entity

type UpdateRegistrationLecturerReq struct {
	UserId string `json:"user_id" validate:"ulid"`

	Id         string `params:"id" validate:"ulid"`
	LecturerId string `json:"lecturer_id" validate:"ulid"`
}

type UpdateRegistrationLecturerResp struct {
	Id string `json:"id"`
}
