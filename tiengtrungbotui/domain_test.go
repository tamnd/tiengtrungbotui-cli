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
	if info.Identity.Binary != "tiengtrungbotui" {
		t.Errorf("Identity.Binary = %q, want tiengtrungbotui", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct{ in, typ, id string }{
		{"tieng-trung-moi-ngay", "series", "tieng-trung-moi-ngay"},
		{"/videos/tieng-trung-moi-ngay/", "series", "videos/tieng-trung-moi-ngay"},
		{"https://" + Host + "/videos/tieng-trung-moi-ngay", "series", "videos/tieng-trung-moi-ngay"},
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
	got, err := Domain{}.Locate("series", "tieng-trung-moi-ngay")
	want := "https://" + Host + "/videos/tieng-trung-moi-ngay"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
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
