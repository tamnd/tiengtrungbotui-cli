// Package tiengtrungbotui is the library behind the ttbt command line:
// the HTTP client, request shaping, and the typed data models for the
// Tieng Trung Bo Tui Vietnamese Chinese-learning site.
//
// The Client here is the spine every command shares. It sets a real
// User-Agent, paces requests so a busy session stays polite, and retries the
// transient failures (429 and 5xx) that any public site throws under load.
package tiengtrungbotui

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// DefaultUserAgent identifies the client to the remote host.
const DefaultUserAgent = "ttbt/dev (+https://github.com/tamnd/tiengtrungbotui-cli)"

// Host is the site this client talks to.
const Host = "tiengtrungbotui.com"

// BaseURL is the root every request is built from.
const BaseURL = "https://" + Host

// Client talks to tiengtrungbotui.com over HTTP.
type Client struct {
	HTTP      *http.Client
	UserAgent string
	Rate      time.Duration
	Retries   int

	mu   sync.Mutex
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
	body, err := c.Get(ctx, BaseURL+"/videos/")
	if err != nil {
		return nil, err
	}
	return parseSeries(string(body))
}

// Episodes fetches a series page and returns all episodes from its JSON-LD data.
func (c *Client) Episodes(ctx context.Context, seriesSlug string) ([]*Episode, error) {
	body, err := c.Get(ctx, BaseURL+"/videos/"+seriesSlug)
	if err != nil {
		return nil, err
	}
	return parseEpisodes(string(body), seriesSlug)
}

// VideoDetail fetches an episode page and returns the full video detail.
func (c *Client) VideoDetail(ctx context.Context, path string) (*Video, error) {
	// path is "series/episode" or full URL
	seriesSlug, episodeSlug, err := splitVideoPath(path)
	if err != nil {
		return nil, err
	}
	body, err := c.Get(ctx, BaseURL+"/videos/"+seriesSlug+"/"+episodeSlug)
	if err != nil {
		return nil, err
	}
	return parseVideo(string(body), seriesSlug, episodeSlug)
}

// AllEpisodes fetches all series then all episodes concurrently (max 4 parallel).
// Results are returned in series order, each series' episodes in their own order.
func (c *Client) AllEpisodes(ctx context.Context) ([]*Episode, error) {
	series, err := c.Series(ctx)
	if err != nil {
		return nil, err
	}

	results := make([][]*Episode, len(series))
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, s := range series {
		wg.Add(1)
		go func(idx int, slug string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			eps, err := c.Episodes(ctx, slug)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
			}
			results[idx] = eps
		}(i, s.Slug)
	}
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	var all []*Episode
	rank := 1
	for _, eps := range results {
		for _, ep := range eps {
			ep.Rank = rank
			rank++
			all = append(all, ep)
		}
	}
	return all, nil
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

func (c *Client) do(ctx context.Context, u string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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
	c.mu.Lock()
	defer c.mu.Unlock()
	if wait := c.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		return 5 * time.Second
	}
	return d
}

// splitVideoPath splits "series/episode" or a full URL into (series, episode).
func splitVideoPath(path string) (seriesSlug, episodeSlug string, err error) {
	// Handle full URLs.
	if strings.HasPrefix(path, "http") {
		u, e := url.Parse(path)
		if e != nil {
			return "", "", e
		}
		path = strings.Trim(u.Path, "/")
	}
	// Strip leading "videos/"
	path = strings.TrimPrefix(path, "videos/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("path must be series/episode, got %q", path)
	}
	return parts[0], parts[1], nil
}
