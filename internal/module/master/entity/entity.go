package entity

type Common struct {
	Id   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}
