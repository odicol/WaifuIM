package models

// RandomImageItems represents a json encoding for every album from the RandomImageResponse.
type RandomImageItems struct {
	Source     string       `json:"source"`
	URL        string       `json:"url"`
	IsAnimated bool         `json:"isAnimated"`
	IsNSFW     bool         `json:"isNsfw"`
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	ByteSize   int          `json:"byteSize"`
	Extension  string       `json:"extension"`
	Tags       []TagItem    `json:"tags"`
	Artists    []ArtistItem `json:"artists"`
}

// RandomImageResponse represents the received format from waifu.im API.
type RandomImageResponse struct {
	Items []RandomImageItems `json:"items"`
}
