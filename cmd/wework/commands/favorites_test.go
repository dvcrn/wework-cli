package commands

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/dvcrn/wework-cli/pkg/wework"
)

func dummyAuth() (*wework.WeWork, error) {
	return nil, nil
}

func TestFormatSpaceTypeName(t *testing.T) {
	tests := []struct {
		spaceType int
		expected  string
	}{
		{spaceType: 0, expected: "Desk"},
		{spaceType: 1, expected: "Office"},
		{spaceType: 2, expected: "Meeting Room"},
		{spaceType: 3, expected: "Event Space"},
		{spaceType: 4, expected: "Type 4"},
		{spaceType: -1, expected: "Type -1"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatSpaceTypeName(tt.spaceType)
			if result != tt.expected {
				t.Errorf("formatSpaceTypeName(%d) = %q, expected %q", tt.spaceType, result, tt.expected)
			}
		})
	}
}

func TestFavoritesCommandStructure(t *testing.T) {
	cmd := NewFavoritesCommand(dummyAuth)

	if cmd.Use != "favorites" {
		t.Errorf("expected command Use to be 'favorites', got %q", cmd.Use)
	}

	expectedAliases := []string{"favorite", "fav", "favs"}
	for _, alias := range expectedAliases {
		if !slices.Contains(cmd.Aliases, alias) {
			t.Errorf("expected alias %q not found in %v", alias, cmd.Aliases)
		}
	}

	subCommands := map[string]bool{
		"list":   false,
		"add":    false,
		"remove": false,
	}

	for _, sub := range cmd.Commands() {
		if _, ok := subCommands[sub.Name()]; ok {
			subCommands[sub.Name()] = true
		}
	}

	for name, found := range subCommands {
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestFavoritesAddValidation(t *testing.T) {
	cmd := newFavoritesAddCommand(dummyAuth)
	cmd.SetArgs([]string{}) // No args, no flags

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error when no arguments or flags are provided")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected error to mention 'required', got: %v", err)
	}
}

func TestFavoritesRemoveValidation(t *testing.T) {
	cmd := newFavoritesRemoveCommand(dummyAuth)
	cmd.SetArgs([]string{}) // No args, no flags

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error when no arguments or flags are provided")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected error to mention 'required', got: %v", err)
	}
}

func TestFavoritesListSpaceTypeValidation(t *testing.T) {
	cmd := newFavoritesListCommand(dummyAuth)
	cmd.SetArgs([]string{"--space-type", "9"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for invalid space type")
	}
	if !strings.Contains(err.Error(), "must be between 0 and 3") {
		t.Errorf("expected error to mention 'must be between 0 and 3', got: %v", err)
	}
}

func TestDeduplicateFavoriteLocations(t *testing.T) {
	locations := []wework.FavoriteLocation{
		{LocationID: "loc-1", LocationName: "Loc 1 Space", SpaceID: "space-1", Hmy: 100},
		{LocationID: "loc-2", LocationName: "Loc 2", Hmy: 200},
		{LocationID: "loc-1", LocationName: "Loc 1 Building", Hmy: 300},
	}

	deduped := deduplicateFavoriteLocations(locations)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 locations after deduplication, got %d", len(deduped))
	}
	if deduped[0].LocationID != "loc-1" || deduped[0].Hmy != 100 {
		t.Errorf("expected first location to be loc-1 with Hmy 100, got %+v", deduped[0])
	}
	if deduped[1].LocationID != "loc-2" {
		t.Errorf("expected second location to be loc-2, got %+v", deduped[1])
	}
}
