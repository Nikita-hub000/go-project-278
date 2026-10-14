package visits

import "time"

type Visit struct {
	ID        int64     `json:"id" db:"id"`
	LinkID    int64     `json:"link_id" db:"link_id" binding:"required"`
	IP        string    `json:"ip" db:"ip" binding:"required"`
	UserAgent string    `json:"user_agent" db:"user_agent" binding:"required"`
	Status    int32     `json:"status" db:"status" binding:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type GetAllVisitsPagination struct {
	Limit  int32
	Offset int32
}
