package topshiriq_dto

type Create struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	CategoryID  int64  `json:"category_id" validate:"required"`
	HududID     int64  `json:"hudud_id" validate:"required"`
	File        string `json:"file"`
	Didline     string `json:"didline" validate:"required"`
}

type Update struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	CategoryID  int64  `json:"category_id" validate:"required"`
	HududID     int64  `json:"hudud_id" validate:"required"`
	File        string `json:"file"`
	Didline     string `json:"didline" validate:"required"`
}
type Filter struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  int64  `json:"category_id"`
	HududID     int64  `json:"hudud_id"`
	Didline     string `json:"didline"`
	Cursor      *int64 `json:"cursor"`
	Direction   string `json:"direction"`
	Limit       int    `json:"limit"`
}
