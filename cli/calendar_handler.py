from icalendar import Calendar, Event
from pathlib import Path


class WeWorkCalendar:
    def __init__(self, wework):
        self.wework = wework

    def generate_calendar(self, output_path):
        cal = Calendar()
        cal.add("prodid", "-//WeWork Calendar//workplaceone//")
        cal.add("version", "2.0")

        bookings = self.wework.get_upcoming_bookings()

        for booking in bookings.bookings:
            event = Event()
            event.add("summary", f"WeWork: {booking.reservable.location.name}")
            event.add("dtstart", booking.starts_at)
            event.add("dtend", booking.ends_at)
            event.add("location", booking.reservable.location.address.line1)
            event.add(
                "description", f"WeWork booking at {booking.reservable.location.name}"
            )
            event.add("uid", booking.uuid)
            cal.add_component(event)

        # Write to file
        Path(output_path).write_bytes(cal.to_ical())
        return output_path
