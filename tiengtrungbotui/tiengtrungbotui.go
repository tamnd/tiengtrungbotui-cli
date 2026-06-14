// Package tiengtrungbotui is the library behind the tiengtrungbotui command line:
// the HTTP client, request shaping, and the typed data models for the
// Tiếng Trung Bỏ Túi Vietnamese Chinese-learning site.
//
// The Client here is the spine every command shares. It sets a real
// User-Agent, paces requests so a busy session stays polite, and retries the
// transient failures (429 and 5xx) that any public site throws under load.
package tiengtrungbotui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// DefaultUserAgent identifies the client to the remote host.
const DefaultUserAgent = "tiengtrungbotui/dev (+https://github.com/tamnd/tiengtrungbotui-cli)"

// Host is the site this client talks to, and the host the URI driver in
// domain.go claims.
const Host = "www.tiengtrungbotui.com"

// BaseURL is the root every request is built from.
const BaseURL = "https://" + Host

var (
	// seriesRe extracts slug + title from the bold-anchor section headers on /videos.
	seriesRe = regexp.MustCompile(`href="(/videos/[a-z0-9-]+)"[^>]*class="[^"]*font-bold[^"]*"[^>]*>(.*?)</a>`)
	// ldjsonRe extracts the content of application/ld+json script tags.
	ldjsonRe = regexp.MustCompile(`<script type="application/ld\+json">(.*?)</script>`)
	// tagRe strips HTML tags.
	tagRe = regexp.MustCompile(`<[^>]+>`)
	// ytIDRe extracts a YouTube video ID from a ytimg.com thumbnail URL.
	ytIDRe = regexp.MustCompile(`/vi/([a-zA-Z0-9_-]{11})/`)
)

// Client talks to tiengtrungbotui.com over HTTP.
type Client struct {
	HTTP      *http.Client
	UserAgent string
	Rate      time.Duration
	Retries   int

	last time.Time
}

// NewClient returns a Client with sensible defaults.
func NewClient() *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 30 * time.Second},
		UserAgent: DefaultUserAgent,
		Rate:      200 * time.Millisecond,
		Retries:   5,
	}
}

// Series fetches the /videos index and returns all video series.
func (c *Client) Series(ctx context.Context) ([]*Series, error) {
	body, err := c.Get(ctx, BaseURL+"/videos")
	if err != nil {
		return nil, err
	}
	html := string(body)
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
		return nil, fmt.Errorf("no series found")
	}
	return out, nil
}

// Episodes fetches a series page and returns all episodes from its JSON-LD data.
func (c *Client) Episodes(ctx context.Context, seriesSlug string) ([]*Episode, error) {
	body, err := c.Get(ctx, BaseURL+"/videos/"+seriesSlug)
	if err != nil {
		return nil, err
	}
	html := string(body)

	var collection struct {
		HasPart []struct {
			Name         string `json:"name"`
			URL          string `json:"url"`
			ThumbnailURL string `json:"thumbnailUrl"`
			Level        string `json:"educationalLevel"`
		} `json:"hasPart"`
	}

	for _, m := range ldjsonRe.FindAllStringSubmatch(html, -1) {
		raw := m[1]
		if err := json.Unmarshal([]byte(raw), &collection); err != nil {
			continue
		}
		if len(collection.HasPart) > 0 {
			break
		}
	}

	if len(collection.HasPart) == 0 {
		return nil, fmt.Errorf("no episodes found for series %q", seriesSlug)
	}

	var out []*Episode
	for i, v := range collection.HasPart {
		ytID := ""
		if m := ytIDRe.FindStringSubmatch(v.ThumbnailURL); m != nil {
			ytID = m[1]
		}
		out = append(out, &Episode{
			Rank:      i + 1,
			Series:    seriesSlug,
			Level:     v.Level,
			Title:     v.Name,
			YouTubeID: ytID,
			URL:       v.URL,
		})
	}
	return out, nil
}

// Get fetches url and returns the response body. It paces and retries according
// to the client's settings.
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	if c.Rate <= 0 {
		return
	}
	if wait := c.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}
