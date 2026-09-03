package wework

import (
	"testing"
	"time"
)

func spaceWithHours(timezone, open, close string) *Workspace {
	return &Workspace{
		OpenTime:  open,
		CloseTime: close,
		Location: Location{
			Name:     "Test Location",
			TimeZone: timezone,
		},
	}
}

func TestBookingWindow(t *testing.T) {
	testCases := []struct {
		name      string
		timezone  string
		openTime  string
		closeTime string
		// date is the booking date as the caller supplies it, in UTC.
		date      string
		wantStart string
		wantEnd   string
		// wantEndUTC is the value actually sent to the API.
		wantEndUTC string
	}{
		{
			// Hub71, Asia/Muscat (UTC+4): 23:59 is not a 30-minute slot.
			name:       "23:59 close in positive offset floors to 23:30",
			timezone:   "Asia/Muscat",
			openTime:   "06:00",
			closeTime:  "23:59",
			date:       "2026-09-11T00:00:00Z",
			wantStart:  "2026-09-11T06:00:00+04:00",
			wantEnd:    "2026-09-11T23:30:00+04:00",
			wantEndUTC: "2026-09-11T19:30:00Z",
		},
		{
			// Paulista 1374: 23:59 in a negative offset also rolls the UTC
			// timestamp onto the next calendar day.
			name:       "23:59 close in negative offset floors to 23:30",
			timezone:   "America/Argentina/Buenos_Aires",
			openTime:   "06:00",
			closeTime:  "23:59",
			date:       "2026-09-11T12:00:00Z",
			wantStart:  "2026-09-11T06:00:00-03:00",
			wantEnd:    "2026-09-11T23:30:00-03:00",
			wantEndUTC: "2026-09-12T02:30:00Z",
		},
		{
			name:       "18:00 close is already on a boundary",
			timezone:   "Asia/Tokyo",
			openTime:   "08:30",
			closeTime:  "18:00",
			date:       "2026-09-11T00:00:00Z",
			wantStart:  "2026-09-11T08:30:00+09:00",
			wantEnd:    "2026-09-11T18:00:00+09:00",
			wantEndUTC: "2026-09-11T09:00:00Z",
		},
		{
			name:       "20:30 close is already on a boundary",
			timezone:   "Europe/Berlin",
			openTime:   "09:00",
			closeTime:  "20:30",
			date:       "2026-09-11T00:00:00Z",
			wantStart:  "2026-09-11T09:00:00+02:00",
			wantEnd:    "2026-09-11T20:30:00+02:00",
			wantEndUTC: "2026-09-11T18:30:00Z",
		},
		{
			// The API does not zero-pad every location's hours.
			name:       "unpadded hours parse like padded ones",
			timezone:   "Asia/Tokyo",
			openTime:   "9:00",
			closeTime:  "18:00",
			date:       "2026-09-11T00:00:00Z",
			wantStart:  "2026-09-11T09:00:00+09:00",
			wantEnd:    "2026-09-11T18:00:00+09:00",
			wantEndUTC: "2026-09-11T09:00:00Z",
		},
		{
			name:       "off-boundary close rounds down not up",
			timezone:   "Asia/Tokyo",
			openTime:   "09:00",
			closeTime:  "18:45",
			date:       "2026-09-11T00:00:00Z",
			wantStart:  "2026-09-11T09:00:00+09:00",
			wantEnd:    "2026-09-11T18:30:00+09:00",
			wantEndUTC: "2026-09-11T09:30:00Z",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			date, err := time.Parse(time.RFC3339, tc.date)
			if err != nil {
				t.Fatalf("failed to parse date: %v", err)
			}

			start, end, err := bookingWindow(date, spaceWithHours(tc.timezone, tc.openTime, tc.closeTime))
			if err != nil {
				t.Fatalf("bookingWindow: %v", err)
			}

			if got := start.Format(time.RFC3339); got != tc.wantStart {
				t.Errorf("start = %s, want %s", got, tc.wantStart)
			}
			if got := end.Format(time.RFC3339); got != tc.wantEnd {
				t.Errorf("end = %s, want %s", got, tc.wantEnd)
			}
			if got := end.UTC().Format("2006-01-02T15:04:05Z"); got != tc.wantEndUTC {
				t.Errorf("end in UTC = %s, want %s", got, tc.wantEndUTC)
			}
			if m := end.Minute(); m != 0 && m != 30 {
				t.Errorf("end minute = %d, want 0 or 30", m)
			}
			if end.Second() != 0 {
				t.Errorf("end second = %d, want 0", end.Second())
			}
			// Rounding down must never extend the booking past the close time.
			if !end.Before(start.AddDate(0, 0, 1)) {
				t.Errorf("end %s spilled onto the next local day", end)
			}
		})
	}
}

func TestBookingWindowRejectsEmptyWindow(t *testing.T) {
	date := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)

	// A close time inside the first slot after opening floors back to the
	// opening time, leaving nothing to book.
	if _, _, err := bookingWindow(date, spaceWithHours("Asia/Tokyo", "09:00", "09:15")); err == nil {
		t.Fatal("expected an error for a window shorter than one slot, got nil")
	}
}

func TestBookingWindowRejectsNilWorkspace(t *testing.T) {
	date := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)

	if _, _, err := bookingWindow(date, nil); err == nil {
		t.Fatal("expected an error for a nil workspace, got nil")
	}
}

func TestBookingWindowRejectsUnknownTimezone(t *testing.T) {
	date := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)

	if _, _, err := bookingWindow(date, spaceWithHours("Not/AZone", "09:00", "18:00")); err == nil {
		t.Fatal("expected an error for an unknown timezone, got nil")
	}
}
