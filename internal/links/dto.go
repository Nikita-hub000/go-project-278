package links

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

type UpdateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

type Link struct {
	ID          int64  `json:"id" db:"id"`
	OriginalURL string `json:"original_url" binding:"required,url"`
	ShortName   string `json:"short_name" binding:"required"`
	ShortURL    string `json:"short_url" binding:"required"`
}

type GenerateShortURLData struct {
	ShortURL  string
	ShortName string
}

type GetAllLinksPagination struct {
	Limit  int32
	Offset int32
}
