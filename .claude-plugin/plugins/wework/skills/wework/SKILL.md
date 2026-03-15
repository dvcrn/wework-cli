---
name: wework
description: Use this skill when a user explicitly wants WeWork actions through the installed `wework` CLI, such as "book a WeWork desk", "list WeWork desks in Tokyo tomorrow", "show my WeWork bookings", "list WeWork locations", or "generate a WeWork calendar export". Do not use it for generic coworking, travel planning, or unrelated calendar requests.
---

# WeWork CLI

Use this skill to operate the installed `wework` binary for real WeWork member tasks. The skill is about using the CLI that is already installed, not reimplementing WeWork API flows.

## Prerequisites

- Confirm the `wework` binary is on `PATH` before relying on it.
- Authentication requires `WEWORK_USERNAME` and `WEWORK_PASSWORD`, or the equivalent `--username` and `--password` flags.
- Prefer the environment variables over inline flags when both are possible.
- Never persist credentials, write them to files, or echo them back into user-visible output.

## When To Use Which Command

- Use `wework locations --city ...` first when the user does not know the location UUID.
- Use `wework desks --date ...` before `wework book` when the user wants to inspect availability.
- Use `wework book --date ...` only when the user explicitly wants to reserve a desk.
- Use `wework bookings` for current or upcoming reservations, and `wework bookings --past` for history.
- Use `wework calendar --calendar-path ...` when the user wants an `.ics` export.
- Use `wework me` for profile or membership context.
- `wework quote` is a stable secondary command when the user wants cost or credit details without creating a booking.
- `wework info` is a stable secondary command for location amenities, hours, and entrance instructions.

## Operating Guidance

- Prefer `--json` when the result will be parsed or summarized programmatically.
- If the location UUID is unknown, resolve it from `locations`, then pass `--location-uuid` into later commands.
- If the user names a city and a specific WeWork, use `--city` plus `--name` for `book`, `quote`, or `info`.
- For `desks`, either `--location-uuid` or `--city` is required.
- For `book` and `quote`, either `--location-uuid` or both `--city` and `--name` are required.
- Calendar exports default to `wework_bookings.ics`; set `--calendar-path` when the output filename matters.

## Date Inputs

- Single date: `2026-03-15`
- Multiple dates: `2026-03-15,2026-03-18,2026-03-20`
- Inclusive range: `2026-03-15~2026-03-19`

Read `references/commands.md` for exact invocation patterns.
Read `references/auth.md` when authentication or secret handling is relevant.
Read `references/booking-flow.md` when booking behavior or failures need diagnosis.

## Failure Handling

- Missing credentials: tell the user the CLI expects `WEWORK_USERNAME` and `WEWORK_PASSWORD` or explicit flags.
- Invalid city or location name: resolve with `wework locations --city ...` first, then retry with the location UUID.
- No desk availability: report that no spaces were returned for that date and location rather than implying a CLI bug.
- Booking or quote failure: distinguish between location resolution errors, availability errors, and API/authentication failures.
- Authentication failure: do not retry blindly with exposed secrets; ask the user to verify credentials or session state.
