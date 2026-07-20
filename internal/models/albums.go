package models

// AlbumItem represents a json encoding for every album from the AlbumsResponse.
type AlbumItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"isDefault"`
	UserID      int    `json:"userId"`
	ImageCount  int    `json:"imageCount"`
}

// AlbumsResponse represents the received format from waifu.im API.
type AlbumsResponse struct {
	Items []AlbumItem `json:"items"`
}

// AlbumMetadata represents the required model to create or update an album.
type AlbumMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
