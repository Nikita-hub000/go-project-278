package links

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ShortName   string `json:"short_name"`
}

type UpdateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ShortName   string `json:"short_name"`
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
