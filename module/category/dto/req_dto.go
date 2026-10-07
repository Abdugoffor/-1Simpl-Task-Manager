package category_dto

type Create struct {
	Name        string `json:"name" validate:"required"`
	Slug        string `json:"slug"`
	Description string `json:"description" validate:"required"`
	IsActive    bool   `json:"is_active"`
}

type Update struct {
	Name        string `json:"name" validate:"required"`
	Slug        string `json:"slug"`
	Description string `json:"description" validate:"required"`
	IsActive    bool   `json:"is_active"`
}

type Filter struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
	Cursor      *int64 `json:"cursor"`
	Direction   string `json:"direction"`
	Limit       int    `json:"limit"`
}
