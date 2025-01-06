from icalendar import Calendar, Event, vDate
from pathlib import Path
from datetime import date

from cli.cli import WeWork


class WeWorkCalendar:
    def __init__(self, wework: WeWork):
        self.wework = wework

    def generate_calendar(self, output_path):
        cal = Calendar()
        cal.add("prodid", "-//WeWork Calendar//workplaceone//")
        cal.add("version", "2.0")

        past_bookings = self.wework.get_past_bookings()
        # get the most recent 10 bookings
        past_bookings = past_bookings.bookings[:10]

        upcoming_bookings = self.wework.get_upcoming_bookings()
        merged_bookings = past_bookings + upcoming_bookings.bookings

        for booking in merged_bookings:
            event = Event()
            event.add("summary", f"WeWork: {booking.reservable.location.name}")

            # Convert to date-only for full day events
            start_date = date(
                booking.starts_at.year, booking.starts_at.month, booking.starts_at.day
            )

            dtstart_with_tzid = f"DTSTART;TZID={booking.timezone}"
            dtend_with_tzid = f"DTEND;TZID={booking.timezone}"

            # For single-day events, use the same date for start and end
            event.add(dtstart_with_tzid, vDate(start_date))
            event.add(dtend_with_tzid, vDate(start_date))
            event.add("TZID", booking.timezone)

            # Disable reminders/alerts
            event.add("X-MICROSOFT-CDO-ALLDAYEVENT", "TRUE")
            event.add("X-MICROSOFT-CDO-BUSYSTATUS", "FREE")
            event.add("X-MICROSOFT-CDO-IMPORTANCE", "1")
            event.add("X-MICROSOFT-DISALLOW-COUNTER", "TRUE")
            event.add("X-APPLE-TRAVEL-ADVISORY-BEHAVIOR", "DISABLED")
            event.add("X-MOZ-LASTACK", "0")

            # Mark as transparent so it doesn't show as busy
            event.add("transp", "TRANSPARENT")
            event.add(
                "url", "https://members.wework.com/workplaceone/content2/your-bookings"
            )

            event.add("location", booking.reservable.location.address.line1)

            # Enhanced description with more booking details
            description = (
                f"WeWork Booking Details:\n"
                f"Location: {booking.reservable.location.name}\n"
                f"Address: {booking.reservable.location.address.line1}\n"
                f"Time: {booking.starts_at.strftime('%I:%M %p')} - {booking.ends_at.strftime('%I:%M %p')}\nBooking ID: {booking.uuid}"
            )
            event.add("description", description)
            event.add("uid", booking.uuid)
            cal.add_component(event)

        # Write to file
        Path(output_path).write_bytes(cal.to_ical())
        return output_path
