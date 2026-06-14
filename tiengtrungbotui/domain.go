package tiengtrungbotui

import (
	"context"
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
			Binary: "tiengtrungbotui",
			Short:  "A command line for Tiếng Trung Bỏ Túi.",
			Long: `A command line for Tiếng Trung Bỏ Túi (tiengtrungbotui.com).

tiengtrungbotui reads public video series and episodes from the site, shapes
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

type seriesIn struct {
	Client *Client `kit:"inject"`
}

type episodesIn struct {
	SeriesSlug string  `kit:"arg" help:"series slug"`
	Client     *Client `kit:"inject"`
}

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
		if err := emit(e); err != nil {
			return err
		}
	}
	return nil
}

func (Domain) Classify(input string) (uriType, id string, err error) {
	id = seriesPath(input)
	if id == "" {
		return "", "", errs.Usage("unrecognized tiengtrungbotui reference: %q", input)
	}
	return "series", id, nil
}

func (Domain) Locate(uriType, id string) (string, error) {
	if uriType != "series" {
		return "", errs.Usage("tiengtrungbotui has no resource type %q", uriType)
	}
	return BaseURL + "/videos/" + strings.Trim(id, "/"), nil
}

func seriesPath(input string) string {
	input = strings.TrimSpace(input)
	if u, err := url.Parse(input); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		return strings.Trim(u.Path, "/")
	}
	return strings.Trim(input, "/")
}
