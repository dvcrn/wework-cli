package wework

import (
	"testing"
)

func TestGetQuoteParameters(t *testing.T) {
	// Test case for Munich - uses inventoryUuid when available
	munichWorkspace := &Workspace{
		UUID:          "6ac970ee-972c-11e8-b488-0ac77f0f6524",
		InventoryUUID: "munich-inventory-uuid",
		Location: Location{
			AccountType: 2,
		},
		Reservable: &WorkspaceReservable{
			KubeId: "131834",
		},
	}

	// Test case for Bangkok - falls back to UUID when no inventoryUuid
	bangkokWorkspace := &Workspace{
		UUID:          "c61971d2-624d-11e9-a390-0e1e2abc3cd0",
		InventoryUUID: "", // No inventory UUID
		Location: Location{
			AccountType: 0,
		},
		Reservable: nil, // No "reservable" object in the response
	}

	// Test case for Tokyo - based on actual dump showing inventoryUuid is used
	tokyoWorkspace := &Workspace{
		UUID:          "eb08c128-e25f-11e8-9de1-0ac77f0f6524",
		InventoryUUID: "52043b70-0bf7-4707-8a6a-b7982dff823b", // From actual Tokyo dump
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
			name:      "Munich - Uses inventoryUuid",
			workspace: munichWorkspace,
			expectedParams: QuoteParameters{
				LocationType: 2,
				SpaceID:      "munich-inventory-uuid",
			},
			expectError: false,
		},
		{
			name:      "Bangkok - Falls back to UUID",
			workspace: bangkokWorkspace,
			expectedParams: QuoteParameters{
				LocationType: 0,
				SpaceID:      "c61971d2-624d-11e9-a390-0e1e2abc3cd0",
			},
			expectError: false,
		},
		{
			name:      "Tokyo - Uses inventoryUuid not KubeId",
			workspace: tokyoWorkspace,
			expectedParams: QuoteParameters{
				LocationType: 4,
				SpaceID:      "52043b70-0bf7-4707-8a6a-b7982dff823b",
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
