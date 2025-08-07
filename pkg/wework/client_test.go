package wework

import (
	"testing"
)

func TestGetQuoteParameters(t *testing.T) {
	// Test case for Munich (System A) - based on get_spaces_munich.folder
	munichWorkspace := &Workspace{
		UUID: "6ac970ee-972c-11e8-b488-0ac77f0f6524",
		Location: Location{
			AccountType: 2,
		},
		Reservable: &WorkspaceReservable{
			KubeId: "131834",
		},
	}

	// Test case for Bangkok (System B) - based on get_spaces_1.folder
	bangkokWorkspace := &Workspace{
		UUID: "c61971d2-624d-11e9-a390-0e1e2abc3cd0",
		Location: Location{
			AccountType: 0,
		},
		Reservable: nil, // No "reservable" object in the response
	}

	// Test case for Tokyo (System A) - based on get_spaces_2.folder
	tokyoWorkspace := &Workspace{
		UUID: "eb08c128-e25f-11e8-9de1-0ac77f0f6524",
		Location: Location{
			AccountType: 4,
		},
		Reservable: &WorkspaceReservable{
			KubeId: "6147",
		},
	}

	testCases := []struct {
		name           string
		workspace      *Workspace
		expectedParams QuoteParameters
		expectError    bool
	}{
		{
			name:      "Munich - System A",
			workspace: munichWorkspace,
			expectedParams: QuoteParameters{
				LocationType: 2,
				SpaceID:      "131834",
			},
			expectError: false,
		},
		{
			name:      "Bangkok - System B",
			workspace: bangkokWorkspace,
			expectedParams: QuoteParameters{
				LocationType: 0,
				SpaceID:      "c61971d2-624d-11e9-a390-0e1e2abc3cd0",
			},
			expectError: false,
		},
		{
			name:      "Tokyo - System A",
			workspace: tokyoWorkspace,
			expectedParams: QuoteParameters{
				LocationType: 4,
				SpaceID:      "6147",
			},
			expectError: false,
		},
		{
			name:        "Nil Workspace",
			workspace:   nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params, err := getQuoteParameters(tc.workspace)

			if (err != nil) != tc.expectError {
				t.Fatalf("getQuoteParameters() error = %v, expectError %v", err, tc.expectError)
			}

			if !tc.expectError {
				if params.LocationType != tc.expectedParams.LocationType {
					t.Errorf("Expected LocationType to be %d, but got %d", tc.expectedParams.LocationType, params.LocationType)
				}
				if params.SpaceID != tc.expectedParams.SpaceID {
					t.Errorf("Expected SpaceID to be '%s', but got '%s'", tc.expectedParams.SpaceID, params.SpaceID)
				}
			}
		})
	}
}
