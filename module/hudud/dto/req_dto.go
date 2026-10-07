package hudud_dto

type Create struct {
	Name     string `json:"name" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type Update struct {
	Name     string `json:"name" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type Filter struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsActive  *bool  `json:"is_active"`
	Cursor    *int64 `json:"cursor"`
	Direction string `json:"direction"`
	Limit     int    `json:"limit"`
}
