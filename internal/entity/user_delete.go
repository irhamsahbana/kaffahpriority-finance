package entity

type DeleteUserReq struct {
	UserID string `json:"user_id" validate:"required,ulid"`

	ID string `json:"id" validate:"required,ulid"`
}
