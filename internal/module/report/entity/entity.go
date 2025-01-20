package entity

type AddStudent struct {
	StudentId *string `json:"student_id" validate:"omitempty,ulid" db:"student_id"`
	Name      *string `json:"name" validate:"omitempty,max=255" db:"name"`
}
