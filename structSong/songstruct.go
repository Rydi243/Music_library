package structSong

type Song struct {
	Group       string `json:"group"`
	Song        string `json:"song"`
	Text        string `json:"text,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
	Link        string `json:"link,omitempty"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}
