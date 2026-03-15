# Command Cookbook

Use these commands with the installed `wework` binary.

## Shared Flags

- `--username` and `--password` are supported on every command, but prefer `WEWORK_USERNAME` and `WEWORK_PASSWORD`.
- `--json` returns structured output and disables spinners.

## Locations

List WeWork locations for a city:

```bash
wework locations --city "Tokyo"
```

JSON output:

```bash
wework locations --city "Tokyo" --json
```

## Desks

Inspect available desks for one location:

```bash
wework desks --date 2026-03-15 --location-uuid LOCATION_UUID
```

Inspect desks by city when the UUID is not known yet:

```bash
wework desks --date 2026-03-15 --city "Tokyo"
```

## Book

Book by known location UUID:

```bash
wework book --date 2026-03-15 --location-uuid LOCATION_UUID
```

Book by city plus fuzzy-matched location name:

```bash
wework book --date 2026-03-15 --city "Tokyo" --name "Shibuya Scramble Square"
```

Book multiple dates:

```bash
wework book --date 2026-03-15,2026-03-18,2026-03-20 --location-uuid LOCATION_UUID
```

Book an inclusive range:

```bash
wework book --date 2026-03-15~2026-03-19 --location-uuid LOCATION_UUID
```

## Bookings

Upcoming bookings:

```bash
wework bookings
```

Past bookings:

```bash
wework bookings --past
```

Past bookings with explicit bounds:

```bash
wework bookings --past --start-date 2026-02-01 --end-date 2026-02-29
```

## Calendar

Generate an ICS export:

```bash
wework calendar --calendar-path ./wework_bookings.ics
```

## Me

Show profile details:

```bash
wework me
```

Include bootstrap data:

```bash
wework me --include-bootstrap
```

## Secondary Commands

Get a quote without booking:

```bash
wework quote --date 2026-03-15 --location-uuid LOCATION_UUID
```

Inspect location amenities and instructions:

```bash
wework info --location-uuid LOCATION_UUID
```
