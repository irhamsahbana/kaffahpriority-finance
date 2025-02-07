package entity

type HRDistributionReq struct {
	UserId string `json:"user_id" validate:"ulid"`

	RegistrationId string  `params:"registration_id" validate:"ulid"`
	HRFeeForMentor float64 `json:"hr_fee_for_mentor" validate:"min=0"`
	HRFeeForHR     float64 `json:"hr_fee_for_hr" validate:"min=0"`
}
