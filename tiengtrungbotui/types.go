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
