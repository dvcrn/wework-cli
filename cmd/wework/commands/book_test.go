package commands

import (
	"strings"
	"testing"

	"github.com/dvcrn/wework-cli/pkg/wework"
)

func TestFindLocationByFuzzyName(t *testing.T) {
	locations := []wework.GeoLocation{
		{Name: "Shibuya Scramble Square", UUID: "uuid1"},
		{Name: "Shinjuku Station", UUID: "uuid2"},
		{Name: "Tokyo Tower", UUID: "uuid3"},
		{Name: "Shibuya Crossing", UUID: "uuid4"},
	}

	tests := []struct {
		name          string
		searchName    string
		expectedUUID  string
		expectError   bool
		errorContains string
	}{
		{
			name:         "exact match",
			searchName:   "Shibuya Scramble Square",
			expectedUUID: "uuid1",
			expectError:  false,
		},
		{
			name:         "fuzzy match single result",
			searchName:   "Shinjuku",
			expectedUUID: "uuid2",
			expectError:  false,
		},
		{
			name:          "multiple matches",
			searchName:    "Shibuya",
			expectedUUID:  "",
			expectError:   true,
			errorContains: "multiple locations found",
		},
		{
			name:          "no match",
			searchName:    "Nonexistent",
			expectedUUID:  "",
			expectError:   true,
			errorContains: "no location found",
		},
		{
			name:         "case insensitive match",
			searchName:   "shibuya scramble square",
			expectedUUID: "uuid1",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid, err := FindLocationByFuzzyName(tt.searchName, locations)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if uuid != tt.expectedUUID {
					t.Errorf("expected UUID %s, got %s", tt.expectedUUID, uuid)
				}
			}
		})
	}
}
