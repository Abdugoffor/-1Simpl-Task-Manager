package topshiriq_dto

type Response struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  int64  `json:"category_id"`
	HududID     int64  `json:"hudud_id"`
	File        string `json:"file"`
	Status      string `json:"status"`
	Didline     string `json:"didline"`	
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type TopshiriqListResponse struct {
	Data       []Response `json:"data"`
	NextCursor *int64     `json:"next_cursor,omitempty"`
	PrevCursor *int64     `json:"prev_cursor,omitempty"`
	HasNext    bool       `json:"has_next"`
	HasPrev    bool       `json:"has_prev"`
}