package tiengtrungbotui

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "tiengtrungbotui" {
		t.Errorf("Scheme = %q, want tiengtrungbotui", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "ttbt" {
		t.Errorf("Identity.Binary = %q, want ttbt", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct{ in, typ, id string }{
		// Bare series slug
		{"tieng-trung-moi-ngay", "series", "tieng-trung-moi-ngay"},
		// Path with /videos/ prefix
		{"/videos/tieng-trung-moi-ngay/", "series", "tieng-trung-moi-ngay"},
		// Full series URL
		{"https://" + Host + "/videos/tieng-trung-moi-ngay", "series", "tieng-trung-moi-ngay"},
		// Episode path "series/episode"
		{"tieng-trung-moi-ngay/hsk-1-2-ep", "episode", "tieng-trung-moi-ngay/hsk-1-2-ep"},
		// Full episode URL
		{"https://" + Host + "/videos/tieng-trung-moi-ngay/hsk-1-2-ep", "episode", "tieng-trung-moi-ngay/hsk-1-2-ep"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestLocate(t *testing.T) {
	cases := []struct {
		uriType, id, want string
	}{
		{"series", "tieng-trung-moi-ngay", BaseURL + "/videos/tieng-trung-moi-ngay"},
		{"episode", "tieng-trung-moi-ngay/hsk-1-2-ep", BaseURL + "/videos/tieng-trung-moi-ngay/hsk-1-2-ep"},
	}
	for _, tc := range cases {
		got, err := Domain{}.Locate(tc.uriType, tc.id)
		if err != nil || got != tc.want {
			t.Errorf("Locate(%q, %q) = (%q, %v), want (%q, nil)", tc.uriType, tc.id, got, err, tc.want)
		}
	}
}

func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}

	got, err := h.ResolveOn("tiengtrungbotui", "tieng-trung-moi-ngay")
	if err != nil || got.String() != "tiengtrungbotui://series/tieng-trung-moi-ngay" {
		t.Errorf("ResolveOn = (%q, %v), want tiengtrungbotui://series/tieng-trung-moi-ngay", got.String(), err)
	}
}
