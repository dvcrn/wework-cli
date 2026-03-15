# Booking Flow And Failure Diagnosis

The CLI uses a two-step booking flow under the hood:

1. Resolve the target location.
2. Fetch available spaces for the requested date.
3. Request a booking quote.
4. Create the booking with the returned quote context.

The user usually only needs `wework book`, but this flow explains the most common failure modes.

## Date Handling

`book` and `quote` accept three formats through `--date`:

- Single date: `YYYY-MM-DD`
- Comma-separated dates: `YYYY-MM-DD,YYYY-MM-DD`
- Inclusive range: `YYYY-MM-DD~YYYY-MM-DD`

The CLI expands multi-date inputs and processes each requested day separately.

## How To Reason About Failures

- Location resolution failure:
  The user gave neither `--location-uuid` nor a valid `--city` and `--name` pair, or the fuzzy location match was ambiguous.

- Availability failure:
  `wework desks` or the booking step returned no spaces for the chosen date and location.

- Ambiguous space failure:
  Multiple spaces were returned when the command expected one target space. Narrow the request by first resolving the exact location and retrying.

- Quote failure:
  The location resolved and spaces existed, but WeWork rejected the quote request. This usually points to upstream validation or auth/session issues.

- Booking failure:
  The quote succeeded but booking creation failed. Report the returned booking status and any returned errors instead of flattening everything into a generic failure.

## Practical Workflow

- If the user is exploring, start with `locations`, then `desks`.
- If they want pricing or credits before committing, run `quote`.
- If they want the reservation created, run `book`.
- If a booking unexpectedly fails, compare `quote` and `book` behavior for the same location and date to isolate whether the problem is availability, quoting, or booking creation.
