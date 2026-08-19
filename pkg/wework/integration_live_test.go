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
		for _, f := range resp.FavoriteLocations {
			t.Logf("  Fav: %+v", f)
		}
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

func TestLiveMarkFavoriteLocation(t *testing.T) {
	ww := liveClient(t)

	// Fetch current favorites
	initial, err := ww.GetFavoriteLocations(0)
	if err != nil {
		t.Fatalf("GetFavoriteLocations: %v", err)
	}

	// Verify we can fetch locations in Tokyo
	locs, err := ww.GetLocationsByGeo("Tokyo")
	if err != nil {
		t.Fatalf("GetLocationsByGeo: %v", err)
	}
	if len(locs.LocationsByGeo) == 0 {
		t.Skip("no locations found in Tokyo")
	}

	isFav := make(map[string]bool)
	for _, f := range initial.FavoriteLocations {
		isFav[f.LocationID] = true
	}

	var targetUUID string
	for _, l := range locs.LocationsByGeo {
		if !isFav[l.UUID] {
			targetUUID = l.UUID
			break
		}
	}
	if targetUUID == "" {
		t.Skip("all locations in Tokyo are already favorited")
	}

	// Add favorite
	addReq := MarkFavoriteLocationRequest{
		LocationID:          targetUUID,
		SpaceType:           0,
		IsDeleted:           false,
		LocationType:        2,
		LocationAccountType: 4,
		PlatformType:        "iOS_APP",
		ApplicationType:     "WorkplaceOne",
	}
	addRes, err := ww.MarkFavoriteLocation(addReq)
	if err != nil {
		t.Fatalf("MarkFavoriteLocation(add): %v", err)
	}

	favID := 0
	if idVal, ok := addRes["FavoriteId"]; ok {
		switch v := idVal.(type) {
		case float64:
			favID = int(v)
		case int:
			favID = v
		}
	}

	if favID > 0 {
		// Clean up by removing
		delReq := MarkFavoriteLocationRequest{
			ID:                  favID,
			SpaceType:           0,
			IsDeleted:           true,
			LocationType:        2,
			LocationAccountType: 4,
			PlatformType:        "iOS_APP",
			ApplicationType:     "WorkplaceOne",
		}
		if _, err := ww.MarkFavoriteLocation(delReq); err != nil {
			t.Fatalf("MarkFavoriteLocation(cleanup): %v", err)
		}
	}
}
