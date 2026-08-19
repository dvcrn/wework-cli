package wework

import (
	"encoding/json"
	"testing"
	"time"
)

// The API reports booking times as location-local wall clock stamped with "Z";
// a Tokyo booking for the location's 09:00-20:00 opening hours arrives as 09:00Z-20:00Z.
const tokyoBookingJSON = `{
	"uuid": "1000001",
	"startsAt": "2026-09-11T09:00:00.000Z",
	"endsAt": "2026-09-11T20:00:00.000Z",
	"modificationDeadline": "2026-09-10T23:00:00.000Z",
	"creditOrder": {"price": "0"},
	"kubeBookingExternalReference": "2000002",
	"reservable": {
		"uuid": "00000000-0000-4000-8000-000000000001",
		"__typename": "SharedWorkspace",
		"location": {
			"uuid": "00000000-0000-4000-8000-000000000002",
			"name": "Shibuya Scramble Square",
			"timeZone": "Asia/Tokyo",
			"accountType": 4,
			"address": {"line1": "Shibuya Scramble Square 39F, 2-24-12 Shibuya", "city": "Tokyo", "country": "JP"}
		}
	}
}`

func decodeTokyoBooking(t *testing.T) *Booking {
	t.Helper()

	var booking Booking
	if err := json.Unmarshal([]byte(tokyoBookingJSON), &booking); err != nil {
		t.Fatalf("failed to decode booking: %v", err)
	}

	w := &WeWork{}
	w.adjustBookingTimezone(&booking)
	return &booking
}

func TestAdjustBookingTimezoneKeepsWallClock(t *testing.T) {
	booking := decodeTokyoBooking(t)

	if got, want := booking.StartsAt.Time.Format(time.RFC3339), "2026-09-11T09:00:00+09:00"; got != want {
		t.Errorf("startsAt = %s, want %s", got, want)
	}
	if got, want := booking.EndsAt.Time.Format(time.RFC3339), "2026-09-11T20:00:00+09:00"; got != want {
		t.Errorf("endsAt = %s, want %s", got, want)
	}
	if got, want := booking.ModificationDeadline.Time.Format(time.RFC3339), "2026-09-10T23:00:00+09:00"; got != want {
		t.Errorf("modificationDeadline = %s, want %s", got, want)
	}
	if got, want := booking.TimeZone, "Asia/Tokyo"; got != want {
		t.Errorf("timezone = %q, want %q", got, want)
	}
}

func TestAdjustBookingTimezoneKeepsZeroTimes(t *testing.T) {
	booking := &Booking{
		Reservable: &SharedWorkspace{
			Location: &SharedWorkspaceLocation{TimeZone: "Asia/Tokyo"},
		},
	}

	w := &WeWork{}
	w.adjustBookingTimezone(booking)

	if !booking.ModificationDeadline.Time.IsZero() {
		t.Errorf("zero modificationDeadline became %s", booking.ModificationDeadline.Time)
	}
}

// The cancel endpoint expects the same wall clock the API reported, so anchoring
// booking times in the location's timezone must not change the payload.
func TestCancelBookingRequestUsesReportedWallClock(t *testing.T) {
	booking := decodeTokyoBooking(t)

	req, err := buildCancelBookingRequest(booking)
	if err != nil {
		t.Fatalf("failed to build cancel request: %v", err)
	}

	payload := req.Body

	if got, want := payload["startTime"], "2026-09-11T09:00:00.000"; got != want {
		t.Errorf("startTime = %v, want %v", got, want)
	}
	if got, want := payload["endTime"], "2026-09-11T20:00:00.000"; got != want {
		t.Errorf("endTime = %v, want %v", got, want)
	}
}
