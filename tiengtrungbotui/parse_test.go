package tiengtrungbotui

import "testing"

const fakeSeriesHTML = `<html><body>
<a href="/videos/tieng-trung-moi-ngay" class="block font-bold text-gray-900">Tiếng Trung mỗi ngày</a>
<a href="/videos/truyen-ke" class="block font-bold text-gray-900">Truyện kể</a>
<a href="/videos/tieng-trung-moi-ngay" class="block font-bold text-gray-900">Duplicate</a>
</body></html>`

const fakeCollectionLD = `<html><body>
<script type="application/ld+json">{"@context":"https://schema.org","@type":"WebSite","name":"TTBT"}</script>
<script type="application/ld+json">{
  "@type":"Collection",
  "hasPart":[
    {
      "name":"HSK 1-2 | Cumtu tieng Trung",
      "url":"https://tiengtrungbotui.com/videos/series/ep1",
      "thumbnailUrl":"https://i.ytimg.com/vi/qMX5RDWk-24/hqdefault.jpg",
      "embedUrl":"https://www.youtube.com/watch?v=qMX5RDWk-24",
      "educationalLevel":"Beginner"
    },
    {
      "name":"HSK 3-4 | Tu vung nang cao",
      "url":"https://tiengtrungbotui.com/videos/series/ep2",
      "thumbnailUrl":"https://i.ytimg.com/vi/ABCDE12345F/hqdefault.jpg",
      "embedUrl":"https://www.youtube.com/watch?v=ABCDE12345F",
      "educationalLevel":"Intermediate"
    }
  ]
}</script>
</body></html>`

const fakeVideoLD = `<html><body>
<script type="application/ld+json">{
  "@type":"VideoObject",
  "url":"https://tiengtrungbotui.com/videos/tieng-trung-moi-ngay/hsk-1-2-ep",
  "name":"HSK 1-2 | Episode title | Luyen nghe",
  "description":"Desc text here.",
  "thumbnailUrl":"https://i.ytimg.com/vi/qMX5RDWk-24/hqdefault.jpg",
  "embedUrl":"https://www.youtube.com/watch?v=qMX5RDWk-24",
  "inLanguage":"zh",
  "learningResourceType":"Listening Practice",
  "educationalLevel":"Beginner"
}</script>
</body></html>`

func TestParseSeries(t *testing.T) {
	series, err := parseSeries(fakeSeriesHTML)
	if err != nil {
		t.Fatal(err)
	}
	if len(series) != 2 {
		t.Fatalf("want 2 series (dedup), got %d", len(series))
	}
	if series[0].Slug != "tieng-trung-moi-ngay" {
		t.Errorf("Slug = %q", series[0].Slug)
	}
	if series[0].Title != "Tiếng Trung mỗi ngày" {
		t.Errorf("Title = %q", series[0].Title)
	}
	if series[0].Rank != 1 {
		t.Errorf("Rank = %d, want 1", series[0].Rank)
	}
	if series[1].Slug != "truyen-ke" {
		t.Errorf("second Slug = %q", series[1].Slug)
	}
}

func TestParseEpisodes(t *testing.T) {
	eps, err := parseEpisodes(fakeCollectionLD, "series")
	if err != nil {
		t.Fatal(err)
	}
	if len(eps) != 2 {
		t.Fatalf("want 2 episodes, got %d", len(eps))
	}
	ep := eps[0]
	if ep.Level != "hsk1-2" {
		t.Errorf("Level = %q, want hsk1-2", ep.Level)
	}
	if ep.YouTubeID != "qMX5RDWk-24" {
		t.Errorf("YouTubeID = %q", ep.YouTubeID)
	}
	if ep.Series != "series" {
		t.Errorf("Series = %q", ep.Series)
	}
	ep2 := eps[1]
	if ep2.Level != "hsk3-4" {
		t.Errorf("ep2 Level = %q, want hsk3-4", ep2.Level)
	}
	if ep2.YouTubeID != "ABCDE12345F" {
		t.Errorf("ep2 YouTubeID = %q", ep2.YouTubeID)
	}
}

func TestParseVideo(t *testing.T) {
	v, err := parseVideo(fakeVideoLD, "tieng-trung-moi-ngay", "hsk-1-2-ep")
	if err != nil {
		t.Fatal(err)
	}
	if v.Series != "tieng-trung-moi-ngay" {
		t.Errorf("Series = %q", v.Series)
	}
	if v.Level != "hsk1-2" {
		t.Errorf("Level = %q, want hsk1-2", v.Level)
	}
	if v.YouTubeID != "qMX5RDWk-24" {
		t.Errorf("YouTubeID = %q", v.YouTubeID)
	}
	if v.Language != "zh" {
		t.Errorf("Language = %q", v.Language)
	}
	if v.ResourceType != "Listening Practice" {
		t.Errorf("ResourceType = %q", v.ResourceType)
	}
}

func TestExtractLevel(t *testing.T) {
	cases := []struct {
		title, level, want string
	}{
		{"HSK 1-2 | Some title", "Beginner", "hsk1-2"},
		{"HSK 3-4 | Another", "Intermediate", "hsk3-4"},
		{"HSK 5 | Advanced topic", "Advanced", "hsk5"},
		{"No HSK prefix here", "Beginner", "beginner"},
		{"No HSK prefix here", "", ""},
		{"HSK 2-3 | Title", "", "hsk2-3"},
	}
	for _, tc := range cases {
		got := extractLevel(tc.title, tc.level)
		if got != tc.want {
			t.Errorf("extractLevel(%q, %q) = %q, want %q", tc.title, tc.level, got, tc.want)
		}
	}
}

func TestExtractYouTubeID(t *testing.T) {
	cases := []struct {
		thumb, embed, want string
	}{
		{"https://i.ytimg.com/vi/qMX5RDWk-24/hqdefault.jpg", "", "qMX5RDWk-24"},
		{"", "https://www.youtube.com/watch?v=ABCDE12345F", "ABCDE12345F"},
		{"https://i.ytimg.com/vi/XXXXXXXXXXX/hqdefault.jpg", "https://www.youtube.com/watch?v=YYYYYYYYYYY", "XXXXXXXXXXX"},
		{"", "", ""},
	}
	for _, tc := range cases {
		got := extractYouTubeID(tc.thumb, tc.embed)
		if got != tc.want {
			t.Errorf("extractYouTubeID(%q, %q) = %q, want %q", tc.thumb, tc.embed, got, tc.want)
		}
	}
}

func TestMatchesCategory(t *testing.T) {
	ep1 := &Episode{Level: "hsk1-2", Title: "HSK 1-2 | Tu vung thi truong"}
	ep2 := &Episode{Level: "hsk3-4", Title: "HSK 3-4 | Ngu phap nang cao"}
	ep3 := &Episode{Level: "beginner", Title: "Lesson for beginners"}

	cases := []struct {
		ep   *Episode
		cat  string
		want bool
	}{
		{ep1, "", true},
		{ep1, "hsk1", true},
		{ep1, "hsk2", false},
		{ep2, "hsk3", true},
		{ep1, "vocab", true},   // title has "tu vung"
		{ep2, "grammar", true}, // title has "ngu phap"
		{ep3, "hsk1", false},
	}
	for _, tc := range cases {
		got := matchesCategory(tc.ep, tc.cat)
		if got != tc.want {
			t.Errorf("matchesCategory(%q, %q) = %v, want %v", tc.ep.Level, tc.cat, got, tc.want)
		}
	}
}
