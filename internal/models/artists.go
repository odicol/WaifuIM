package models

// ArtistItem represents a json encoding for every album from the ArtistsResponse.
type ArtistItem struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Patreon string `json:"patreon"`
	Pixiv   string `json:"pixiv"`
	Twitter string `json:"twitter"`
}

// ArtistsResponse represents the received format from waifu.im API.
type ArtistsResponse struct {
	Items []ArtistItem `json:"items"`
}
