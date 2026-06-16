package tiengtrungbotui

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

func init() { kit.Register(Domain{}) }

// Domain is the tiengtrungbotui driver.
type Domain struct{}

func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "tiengtrungbotui",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "ttbt",
			Short:  "A command line for Tieng Trung Bo Tui.",
			Long: `A command line for Tieng Trung Bo Tui (tiengtrungbotui.com).

ttbt reads public video series and episodes from the site, shapes
them into clean records, and prints output that pipes into the rest of your
tools. No API key, nothing to run alongside it.`,
			Site: Host,
			Repo: "https://github.com/tamnd/tiengtrungbotui-cli",
		},
	}
}

func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{Name: "series", Group: "read", List: true,
		Summary: "List all video series"}, listSeries)

	kit.Handle(app, kit.OpMeta{Name: "episodes", Group: "read", List: true,
		Summary: "List episodes in a series",
		Args:    []kit.Arg{{Name: "series", Help: "series slug"}}}, listEpisodes)

	kit.Handle(app, kit.OpMeta{Name: "video", Group: "read", Single: true,
		Summary: "Fetch full detail for one episode",
		Args:    []kit.Arg{{Name: "path", Help: "series/episode path or full URL"}}}, getVideo)

	kit.Handle(app, kit.OpMeta{Name: "lesson", Group: "read", Single: true,
		Summary: "Fetch one lesson (alias for video)",
		Args:    []kit.Arg{{Name: "path", Help: "series/episode path or full URL"}}}, getVideo)

	kit.Handle(app, kit.OpMeta{Name: "list", Group: "read", List: true,
		Summary: "List all episodes across all series"}, listAll)

	kit.Handle(app, kit.OpMeta{Name: "search", Group: "read", List: true,
		Summary: "Search episodes by title",
		Args:    []kit.Arg{{Name: "query", Help: "search term"}}}, searchEpisodes)

	kit.Handle(app, kit.OpMeta{Name: "word", Group: "read", List: true,
		Summary: "Find episodes mentioning a word",
		Args:    []kit.Arg{{Name: "word", Help: "Chinese word or Pinyin"}}}, wordLookup)

	kit.Handle(app, kit.OpMeta{Name: "export", Group: "read", List: true,
		Summary: "Export all episodes as JSONL"}, exportAll)

	kit.Handle(app, kit.OpMeta{Name: "info", Group: "read", Single: true,
		Summary: "Print site statistics"}, siteInfo)
}

func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := NewClient()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.HTTP.Timeout = cfg.Timeout
	}
	return c, nil
}

// --- input structs ---

type seriesIn struct {
	Client *Client `kit:"inject"`
}

type episodesIn struct {
	SeriesSlug string  `kit:"arg"  help:"series slug"`
	Level      string  `kit:"flag,name=level" help:"filter by HSK level (e.g. hsk1-2)"`
	Client     *Client `kit:"inject"`
}

type videoIn struct {
	Path   string  `kit:"arg"  help:"series/episode path or full URL"`
	Client *Client `kit:"inject"`
}

type listIn struct {
	Category string  `kit:"flag,name=category" help:"filter: grammar|vocab|hsk1|hsk2|hsk3|hsk4|hsk5|hsk6"`
	Client   *Client `kit:"inject"`
}

type searchIn struct {
	Query  string  `kit:"arg"  help:"search term"`
	Client *Client `kit:"inject"`
}

type wordIn struct {
	Word   string  `kit:"arg"  help:"Chinese word or Pinyin"`
	Client *Client `kit:"inject"`
}

type exportIn struct {
	Client *Client `kit:"inject"`
}

type infoIn struct {
	Client *Client `kit:"inject"`
}

// --- handlers ---

func listSeries(ctx context.Context, in seriesIn, emit func(*Series) error) error {
	items, err := in.Client.Series(ctx)
	if err != nil {
		return err
	}
	for _, s := range items {
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

func listEpisodes(ctx context.Context, in episodesIn, emit func(*Episode) error) error {
	items, err := in.Client.Episodes(ctx, in.SeriesSlug)
	if err != nil {
		return err
	}
	for _, e := range items {
		if in.Level != "" && !strings.HasPrefix(e.Level, strings.ToLower(in.Level)) {
			continue
		}
		if err := emit(e); err != nil {
			return err
		}
	}
	return nil
}

func getVideo(ctx context.Context, in videoIn, emit func(*Video) error) error {
	v, err := in.Client.VideoDetail(ctx, in.Path)
	if err != nil {
		return err
	}
	return emit(v)
}

func listAll(ctx context.Context, in listIn, emit func(*Episode) error) error {
	all, err := in.Client.AllEpisodes(ctx)
	if err != nil {
		return err
	}
	for _, ep := range all {
		if !matchesCategory(ep, in.Category) {
			continue
		}
		if err := emit(ep); err != nil {
			return err
		}
	}
	return nil
}

func searchEpisodes(ctx context.Context, in searchIn, emit func(*Episode) error) error {
	all, err := in.Client.AllEpisodes(ctx)
	if err != nil {
		return err
	}
	q := strings.ToLower(in.Query)
	for _, ep := range all {
		if strings.Contains(strings.ToLower(ep.Title), q) {
			if err := emit(ep); err != nil {
				return err
			}
		}
	}
	return nil
}

func wordLookup(ctx context.Context, in wordIn, emit func(*Episode) error) error {
	all, err := in.Client.AllEpisodes(ctx)
	if err != nil {
		return err
	}
	q := strings.ToLower(in.Word)
	for _, ep := range all {
		if strings.Contains(strings.ToLower(ep.Title), q) {
			if err := emit(ep); err != nil {
				return err
			}
		}
	}
	return nil
}

func exportAll(ctx context.Context, in exportIn, emit func(*Episode) error) error {
	all, err := in.Client.AllEpisodes(ctx)
	if err != nil {
		return err
	}
	for _, ep := range all {
		if err := emit(ep); err != nil {
			return err
		}
	}
	return nil
}

func siteInfo(ctx context.Context, in infoIn, emit func(*Stats) error) error {
	series, err := in.Client.Series(ctx)
	if err != nil {
		return err
	}
	all, err := in.Client.AllEpisodes(ctx)
	if err != nil {
		return err
	}
	levels := map[string]int{}
	for _, ep := range all {
		if ep.Level != "" {
			levels[ep.Level]++
		}
	}
	return emit(&Stats{
		Series:   len(series),
		Episodes: len(all),
		Levels:   levels,
	})
}

// Classify implements the kit domain URI driver.
func (Domain) Classify(input string) (uriType, id string, err error) {
	clean := strings.TrimSpace(input)
	// Full URL
	if strings.HasPrefix(clean, "http") {
		u, e := url.Parse(clean)
		if e != nil {
			return "", "", errs.Usage("invalid URL: %v", e)
		}
		path := strings.Trim(u.Path, "/")
		path = strings.TrimPrefix(path, "videos/")
		if strings.Contains(path, "/") {
			return "episode", path, nil
		}
		return "series", path, nil
	}
	// "series/episode" path
	path := strings.Trim(clean, "/")
	path = strings.TrimPrefix(path, "videos/")
	if strings.Contains(path, "/") {
		return "episode", path, nil
	}
	if path == "" {
		return "", "", errs.Usage("unrecognized tiengtrungbotui reference: %q", input)
	}
	return "series", path, nil
}

// Locate implements the kit domain URI driver.
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "series":
		return fmt.Sprintf("%s/videos/%s", BaseURL, strings.Trim(id, "/")), nil
	case "episode":
		return fmt.Sprintf("%s/videos/%s", BaseURL, strings.Trim(id, "/")), nil
	default:
		return "", errs.Usage("tiengtrungbotui has no resource type %q", uriType)
	}
}
