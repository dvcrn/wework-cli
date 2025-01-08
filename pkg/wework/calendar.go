package wework

import (
	"fmt"
	ics "github.com/arran4/golang-ical"
	"os"
)

type WeWorkCalendar struct {
	client *WeWork
}

func NewWeWorkCalendar(client *WeWork) *WeWorkCalendar {
	return &WeWorkCalendar{
		client: client,
	}
}

func (w *WeWorkCalendar) GenerateCalendar(outputPath string) error {
	cal := ics.NewCalendar()
	cal.SetProductId("-//WeWork Calendar//workplaceone//")
	cal.SetVersion("2.0")

	// Get past and upcoming bookings
	pastBookings, err := w.client.GetPastBookings()
	if err != nil {
		return fmt.Errorf("failed to get past bookings: %v", err)
	}

	upcomingBookings, err := w.client.GetUpcomingBookings()
	if err != nil {
		return fmt.Errorf("failed to get upcoming bookings: %v", err)
	}

	// Limit past bookings to 10 most recent
	if len(pastBookings) > 10 {
		pastBookings = pastBookings[:10]
	}

	// Merge bookings
	allBookings := append(pastBookings, upcomingBookings...)

	// Create events for each booking
	for _, booking := range allBookings {
		event := cal.AddEvent(booking.UUID)
		event.SetSummary(fmt.Sprintf("WeWork: %s", booking.Reservable.Location.Name))

		event.SetProperty(ics.ComponentProperty("DTSTART;TZID="+booking.Reservable.Location.TimeZone),
			booking.StartsAt.Format("20060102"))
		event.SetProperty(ics.ComponentProperty("DTEND;TZID="+booking.Reservable.Location.TimeZone),
			booking.StartsAt.Format("20060102"))

		event.SetProperty(ics.ComponentProperty("TZID"), booking.Reservable.Location.TimeZone)

		// Set Microsoft and Apple specific properties
		event.SetProperty("X-MICROSOFT-CDO-ALLDAYEVENT", "TRUE")
		event.SetProperty("X-MICROSOFT-CDO-BUSYSTATUS", "FREE")
		event.SetProperty("X-MICROSOFT-CDO-IMPORTANCE", "1")
		event.SetProperty("X-MICROSOFT-DISALLOW-COUNTER", "TRUE")
		event.SetProperty("X-APPLE-TRAVEL-ADVISORY-BEHAVIOR", "DISABLED")
		event.SetProperty("X-MOZ-LASTACK", "0")

		// Set transparency and URL
		event.SetProperty("TRANSP", "TRANSPARENT")
		event.SetProperty("URL", "https://members.wework.com/workplaceone/content2/your-bookings")

		// Set location
		event.SetLocation(booking.Reservable.Location.Address.Line1)

		// Set description
		description := fmt.Sprintf(
			"WeWork Booking Details:\nLocation: %s\nAddress: %s\nTime: %s - %s\nBooking ID: %s",
			booking.Reservable.Location.Name,
			booking.Reservable.Location.Address.Line1,
			booking.StartsAt.Format("03:04 PM"),
			booking.EndsAt.Format("03:04 PM"),
			booking.UUID,
		)
		event.SetDescription(description)
	}

	// Write to file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer f.Close()

	return cal.SerializeTo(f)
}
