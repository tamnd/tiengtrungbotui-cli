package tiengtrungbotui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := NewClient()
	c.Rate = 0

	body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "ok" {
		t.Errorf("body = %q, want %q", body, "ok")
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("recovered"))
	}))
	defer srv.Close()

	c := NewClient()
	c.Rate = 0
	c.Retries = 5

	start := time.Now()
	body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "recovered" {
		t.Errorf("body = %q after retries", body)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestGetNonRetriableOn404(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClient()
	c.Rate = 0
	c.Retries = 5

	_, err := c.Get(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if hits != 1 {
		t.Errorf("expected exactly 1 attempt on 404, got %d", hits)
	}
}

func TestPace(t *testing.T) {
	var count int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := NewClient()
	c.Rate = 100 * time.Millisecond

	start := time.Now()
	_, _ = c.Get(context.Background(), srv.URL)
	_, _ = c.Get(context.Background(), srv.URL)
	elapsed := time.Since(start)

	if elapsed < 90*time.Millisecond {
		t.Errorf("second request arrived too fast: %v (want >= 100ms total)", elapsed)
	}
}
