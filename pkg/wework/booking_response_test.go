package wework

import (
	"encoding/json"
	"strings"
	"testing"
)

// The booking endpoint answers 200 even when it declines the reservation, so
// these bodies are what a caller would otherwise accept as a success.
func TestValidateBookingResponse(t *testing.T) {
	testCases := []struct {
		name string
		body string
		// wantErrContains are substrings the error must surface so a caller can
		// see why WeWork rejected the booking.
		wantErrContains []string
	}{
		{
			name: "confirmed booking",
			body: `{"BookingStatus":"BookingSuccess","Errors":[],"ReservationID":"RES-12345"}`,
		},
		{
			name:            "failed status",
			body:            `{"BookingStatus":"BookingFailed","Errors":[],"ReservationID":""}`,
			wantErrContains: []string{"BookingFailed"},
		},
		{
			name:            "pending status",
			body:            `{"BookingStatus":"BookingPending","Errors":[],"ReservationID":"RES-12345"}`,
			wantErrContains: []string{"BookingPending"},
		},
		{
			name: "errors reported alongside a success status",
			body: `{"BookingStatus":"BookingSuccess","Errors":["PlaceOrder","Unable to save^End Time value must be followed 30 minute slots."],"ReservationID":"RES-12345"}`,
			wantErrContains: []string{
				"PlaceOrder",
				"End Time value must be followed 30 minute slots.",
			},
		},
		{
			name:            "success status without a reservation ID",
			body:            `{"BookingStatus":"BookingSuccess","Errors":[],"ReservationID":""}`,
			wantErrContains: []string{"no reservation ID"},
		},
		{
			name:            "empty body",
			body:            `{}`,
			wantErrContains: []string{"unknown"},
		},
	}

	t.Run("nil response", func(t *testing.T) {
		if err := validateBookingResponse(nil); err == nil {
			t.Fatal("validateBookingResponse(nil) = nil, want an error")
		}
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result BookingResponse
			if err := json.Unmarshal([]byte(tc.body), &result); err != nil {
				t.Fatalf("failed to decode body: %v", err)
			}

			err := validateBookingResponse(&result)

			if len(tc.wantErrContains) == 0 {
				if err != nil {
					t.Fatalf("validateBookingResponse = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatal("validateBookingResponse = nil, want an error")
			}
			for _, want := range tc.wantErrContains {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err, want)
				}
			}
		})
	}
}
