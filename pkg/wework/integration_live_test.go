package wework

import (
	"context"
	"os"
	"testing"
)

// liveClient authenticates against the real WeWork API using WEWORK_USERNAME and
// WEWORK_PASSWORD. Tests that call it skip when those are not set, so the package
// test suite stays hermetic by default. Run with:
//
//	WEWORK_USERNAME=... WEWORK_PASSWORD=... go test ./pkg/wework -run Live -v
func liveClient(t *testing.T) *WeWork {
	t.Helper()
	user := os.Getenv("WEWORK_USERNAME")
	pass := os.Getenv("WEWORK_PASSWORD")
	if user == "" || pass == "" {
		t.Skip("set WEWORK_USERNAME and WEWORK_PASSWORD to run live integration tests")
	}
	auth, err := NewWeWorkAuth(user, pass)
	if err != nil {
		t.Fatalf("NewWeWorkAuth: %v", err)
	}
	login, _, err := auth.Authenticate()
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	return NewWeWork(login.A0token)
}

func TestLiveGetFavoriteLocations(t *testing.T) {
	ww := liveClient(t)
	// spaceType must be 0-3; the API rejects other values with a 400.
	for st := 0; st <= 3; st++ {
		resp, err := ww.GetFavoriteLocations(st)
		if err != nil {
			t.Fatalf("GetFavoriteLocations(%d): %v", st, err)
		}
		t.Logf("spaceType=%d favorites=%d recents=%d", st, len(resp.FavoriteLocations), len(resp.RecentLocations))
	}
}

func TestLiveGetPrintQueue(t *testing.T) {
	ww := liveClient(t)
	resp, err := ww.GetPrintQueue(context.Background(), "")
	if err != nil {
		t.Fatalf("GetPrintQueue: %v", err)
	}
	t.Logf("print queue jobs=%d totalElements=%d", len(resp.Content), resp.Page.TotalElements)
}
