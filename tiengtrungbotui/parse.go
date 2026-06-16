package tiengtrungbotui

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	// seriesRe extracts slug + title from bold-anchor section headers on /videos.
	seriesRe = regexp.MustCompile(`href="(/videos/[a-z0-9-]+)"[^>]*class="[^"]*font-bold[^"]*"[^>]*>(.*?)</a>`)
	// ldjsonRe extracts the content of application/ld+json script tags.
	ldjsonRe = regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)
	// tagRe strips HTML tags.
	tagRe = regexp.MustCompile(`<[^>]+>`)
	// ytIDThumbRe extracts a YouTube video ID from a ytimg.com thumbnail URL.
	ytIDThumbRe = regexp.MustCompile(`/vi/([a-zA-Z0-9_-]{11})/`)
	// ytIDEmbedRe extracts YouTube ID from an embed or watch URL.
	ytIDEmbedRe = regexp.MustCompile(`(?:youtube\.com/(?:embed/|watch\?v=)|youtu\.be/)([a-zA-Z0-9_-]{11})`)
	// hskLevelRe extracts "HSK N-M" or "HSK N" from a title prefix.
	hskLevelRe = regexp.MustCompile(`^HSK\s+([\d][\d-]*)`)
)

// parseSeries parses bold-anchor series links from the /videos/ HTML page.
func parseSeries(html string) ([]*Series, error) {
	var out []*Series
	seen := map[string]bool{}
	rank := 1
	for _, m := range seriesRe.FindAllStringSubmatch(html, -1) {
		path := m[1]
		if seen[path] {
			continue
		}
		seen[path] = true
		slug := strings.TrimPrefix(path, "/videos/")
		title := strings.TrimSpace(tagRe.ReplaceAllString(m[2], ""))
		out = append(out, &Series{
			Rank:  rank,
			Slug:  slug,
			Title: title,
			URL:   BaseURL + path,
		})
		rank++
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no series found at %s/videos/", BaseURL)
	}
	return out, nil
}

// collectionLD is the JSON-LD Collection type used on series pages.
type collectionLD struct {
	Type    string `json:"@type"`
	HasPart []struct {
		Name         string `json:"name"`
		URL          string `json:"url"`
		ThumbnailURL string `json:"thumbnailUrl"`
		EmbedURL     string `json:"embedUrl"`
		Level        string `json:"educationalLevel"`
	} `json:"hasPart"`
}

// videoLD is the JSON-LD VideoObject type used on episode pages.
type videoLD struct {
	Type         string `json:"@type"`
	URL          string `json:"url"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnailUrl"`
	EmbedURL     string `json:"embedUrl"`
	Language     string `json:"inLanguage"`
	ResourceType string `json:"learningResourceType"`
	Level        string `json:"educationalLevel"`
}

// parseEpisodes parses the hasPart array from a series page JSON-LD block.
func parseEpisodes(html, seriesSlug string) ([]*Episode, error) {
	for _, m := range ldjsonRe.FindAllStringSubmatch(html, -1) {
		var col collectionLD
		if err := json.Unmarshal([]byte(m[1]), &col); err != nil {
			continue
		}
		if len(col.HasPart) == 0 {
			continue
		}
		var out []*Episode
		for i, v := range col.HasPart {
			ytID := extractYouTubeID(v.ThumbnailURL, v.EmbedURL)
			level := extractLevel(v.Name, v.Level)
			out = append(out, &Episode{
				Rank:      i + 1,
				Series:    seriesSlug,
				Level:     level,
				Title:     v.Name,
				YouTubeID: ytID,
				URL:       v.URL,
			})
		}
		return out, nil
	}
	return nil, fmt.Errorf("no episodes found for series %q", seriesSlug)
}

// parseVideo parses the VideoObject JSON-LD from an episode page.
func parseVideo(html, seriesSlug, episodeSlug string) (*Video, error) {
	for _, m := range ldjsonRe.FindAllStringSubmatch(html, -1) {
		var v videoLD
		if err := json.Unmarshal([]byte(m[1]), &v); err != nil {
			continue
		}
		if v.Type != "VideoObject" || v.Name == "" {
			continue
		}
		ytID := extractYouTubeID(v.ThumbnailURL, v.EmbedURL)
		level := extractLevel(v.Name, v.Level)
		return &Video{
			Series:       seriesSlug,
			Slug:         episodeSlug,
			Level:        level,
			Title:        v.Name,
			Description:  v.Description,
			YouTubeID:    ytID,
			EmbedURL:     v.EmbedURL,
			Language:     v.Language,
			ResourceType: v.ResourceType,
			URL:          v.URL,
		}, nil
	}
	return nil, fmt.Errorf("no video found at %s/videos/%s/%s", BaseURL, seriesSlug, episodeSlug)
}

// extractYouTubeID extracts a YouTube video ID from thumbnail URL or embed URL.
func extractYouTubeID(thumbnailURL, embedURL string) string {
	if thumbnailURL != "" {
		if m := ytIDThumbRe.FindStringSubmatch(thumbnailURL); m != nil {
			return m[1]
		}
	}
	if embedURL != "" {
		if m := ytIDEmbedRe.FindStringSubmatch(embedURL); m != nil {
			return m[1]
		}
	}
	return ""
}

// extractLevel derives a normalized HSK level string from an episode title and
// educationalLevel field.
//
// Priority:
//  1. Title prefix "HSK N-M" -> "hskN-M"
//  2. educationalLevel field -> lowercase
func extractLevel(title, educationalLevel string) string {
	if m := hskLevelRe.FindStringSubmatch(title); m != nil {
		return "hsk" + m[1]
	}
	if educationalLevel != "" {
		return strings.ToLower(educationalLevel)
	}
	return ""
}

// matchesCategory returns true when the episode matches the given category filter.
func matchesCategory(ep *Episode, category string) bool {
	if category == "" {
		return true
	}
	cat := strings.ToLower(category)
	switch cat {
	case "vocab":
		return strings.Contains(strings.ToLower(ep.Title), "tu vung") ||
			strings.Contains(strings.ToLower(ep.Title), "từ vựng")
	case "grammar":
		return strings.Contains(strings.ToLower(ep.Title), "ngu phap") ||
			strings.Contains(strings.ToLower(ep.Title), "ngữ pháp")
	default:
		// "hsk1", "hsk2", etc. -- match level prefix
		return strings.HasPrefix(ep.Level, cat)
	}
}
