package tiengtrungbotui

// Series is one video playlist on tiengtrungbotui.com.
type Series struct {
	Rank  int    `json:"rank"  csv:"rank"  tsv:"rank"`
	Slug  string `json:"slug"  csv:"slug"  tsv:"slug"  kit:"id"`
	Title string `json:"title" csv:"title" tsv:"title"`
	URL   string `json:"url"   csv:"url"   tsv:"url"`
}

// Episode is one video inside a series.
type Episode struct {
	Rank      int    `json:"rank"       csv:"rank"       tsv:"rank"`
	Series    string `json:"series"     csv:"series"     tsv:"series"`
	Level     string `json:"level"      csv:"level"      tsv:"level"`
	Title     string `json:"title"      csv:"title"      tsv:"title"`
	YouTubeID string `json:"youtube_id" csv:"youtube_id" tsv:"youtube_id"`
	URL       string `json:"url"        csv:"url"        tsv:"url"`
}

// Video is the full detail for one episode.
type Video struct {
	Series       string `json:"series"        csv:"series"        tsv:"series"`
	Slug         string `json:"slug"          csv:"slug"          tsv:"slug"`
	Level        string `json:"level"         csv:"level"         tsv:"level"`
	Title        string `json:"title"         csv:"title"         tsv:"title"`
	Description  string `json:"description"   csv:"description"   tsv:"description"`
	YouTubeID    string `json:"youtube_id"    csv:"youtube_id"    tsv:"youtube_id"`
	EmbedURL     string `json:"embed_url"     csv:"embed_url"     tsv:"embed_url"`
	Language     string `json:"language"      csv:"language"      tsv:"language"`
	ResourceType string `json:"resource_type" csv:"resource_type" tsv:"resource_type"`
	URL          string `json:"url"           csv:"url"           tsv:"url"`
}

// Stats holds site-wide statistics for ttbt info.
type Stats struct {
	Series   int            `json:"series"`
	Episodes int            `json:"episodes"`
	Levels   map[string]int `json:"levels"`
}
