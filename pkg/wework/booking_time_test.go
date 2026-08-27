package wework

import (
	"encoding/json"
	"testing"
	"time"
)

// The API reports booking times as location-local wall clock stamped with "Z";
// a Tokyo booking for the location's 09:00-20:00 opening hours arrives as 09:00Z-20:00Z.
const tokyoBookingJSON = `{
	"bookingId": "1000001",
	"kubeBookingExternalReference": "2000002",
	"franchiseBookingExtReference": "00000000-0000-4000-8000-000000000003",
	"isFranchiseBooking": true,
	"startDate": "2026-09-11T09:00:00.000Z",
	"endDate": "2026-09-11T20:00:00.000Z",
	"modificationDeadlineTime": "2026-09-10T23:00:00.000Z",
	"creditCost": 0,
	"spaceId": "00000000-0000-4000-8000-000000000001",
	"spaceExternalReference": "00000000-0000-4000-8000-000000000001",
	"spaceName": "Daily Desk",
	"location": {
		"id": "00000000-0000-4000-8000-000000000002",
		"name": "Shibuya Scramble Square",
		"timeZoneIana": "Asia/Tokyo",
		"sourceType": 4,
		"address": {"line1": "Shibuya Scramble Square 39F, 2-24-12 Shibuya", "city": "Tokyo", "country": "JP"}
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
		Location: &BookingLocation{TimeZone: "Asia/Tokyo"},
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

// Franchise bookings are cancelled by their kube reference and space external
// reference rather than by bookingId/spaceId.
func TestCancelBookingRequestUsesFranchiseIdentifiers(t *testing.T) {
	booking := decodeTokyoBooking(t)

	req, err := buildCancelBookingRequest(booking)
	if err != nil {
		t.Fatalf("failed to build cancel request: %v", err)
	}

	if got, want := req.Body["bookingId"], "2000002"; got != want {
		t.Errorf("bookingId = %v, want %v", got, want)
	}
	if got, want := req.Body["reservableId"], "00000000-0000-4000-8000-000000000001"; got != want {
		t.Errorf("reservableId = %v, want %v", got, want)
	}
	if got, want := req.Body["bookingLocationType"], 4; got != want {
		t.Errorf("bookingLocationType = %v, want %v", got, want)
	}

	booking.IsFranchiseBooking = false
	req, err = buildCancelBookingRequest(booking)
	if err != nil {
		t.Fatalf("failed to build cancel request: %v", err)
	}
	if got, want := req.Body["bookingId"], "1000001"; got != want {
		t.Errorf("non-franchise bookingId = %v, want %v", got, want)
	}
}
