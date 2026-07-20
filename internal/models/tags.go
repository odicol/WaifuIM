package models

// TagItem represents a json encoding for every album from the TagsResponse.
type TagItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// TagsResponse represents the received format from waifu.im API.
type TagsResponse struct {
	Items []TagItem `json:"items"`
}
