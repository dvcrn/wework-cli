# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

- Build: `go build ./cmd/wework`
- Install locally: `go install ./cmd/wework`
- Run: `go run ./cmd/wework/main.go`
- Format code: `go fmt ./...`
- Lint: `golint ./...`
- Vet: `go vet ./...`
- Run tests: `go test ./...` (tests currently missing)

## Architecture Overview

This WeWork CLI tool uses a clean separation between the CLI interface (`cmd/wework`) and core library (`pkg/wework`). The architecture follows a command pattern using Cobra, with lazy authentication passed as callbacks to commands.

### Authentication Flow

The authentication system implements OAuth2 with PKCE via Auth0:
1. Commands receive an authentication callback from main.go
2. `pkg/wework/auth.go` handles the OAuth2 flow with Auth0
3. Tokens are extracted and used for WeWork API authentication
4. User UUID is parsed from JWT and added to request headers
5. Cookie jar maintains session state across requests

### Key Components

- **cmd/wework/main.go**: Entry point with authentication callback pattern
- **pkg/wework/client.go**: Main API client with all WeWork endpoints
- **pkg/wework/baseclient.go**: HTTP client with browser-like headers and compression
- **pkg/wework/types.go**: All data structures and API response types
- **cmd/wework/commands/**: Individual command implementations following Cobra pattern

### API Integration

- Base URL: `https://members.wework.com/workplaceone/api/`
- Authentication: Bearer token in both `Authorization` and `WeWorkAuth` headers
- User UUID from JWT required in `wework-user-uuid` header
- Two-step booking process: quote generation → booking creation

## Adding New Features

### Adding a New Command
1. Create file in `/cmd/wework/commands/` following existing pattern
2. Register command in `main.go` with authentication callback
3. Use consistent error handling: `return fmt.Errorf("context: %w", err)`

### Adding API Endpoints
1. Add method to `WeWork` struct in `pkg/wework/client.go`
2. Define request/response types in `pkg/wework/types.go`
3. Follow existing patterns for error handling and response parsing

### Date Handling
The CLI supports multiple date formats:
- Single date: `2023-06-01`
- Date range: `2023-06-01~2023-06-05`
- Multiple dates: `2023-06-01,2023-06-03,2023-06-05`

Parse dates using the existing pattern in `book.go` command.

## Important Patterns

### Environment Variables
- `WEWORK_USERNAME` and `WEWORK_PASSWORD` for authentication
- Check environment first, then fall back to flags
- Never persist credentials

### Error Handling
- Always wrap errors with context using `fmt.Errorf`
- Use `WeWorkLoginError` for authentication failures
- Return meaningful error messages to users

### Timezone Handling
- All bookings use location-specific timezones
- Use `time.LoadLocation()` with location timezone from API
- Calendar generation includes proper timezone data

## Booking API Details

The booking system requires a two-step process with specific field mappings and headers:

### Two-Step Booking Process
1. **Get Quote** (`/common-booking/quote`): Validates availability and returns pricing
2. **Create Booking** (`/common-booking/`): Confirms the reservation

### Required Headers
- `Request-Source`: Must use iOS app format: `com.wework.ondemand/WorkplaceOne/Prod/iOS/2.70.1(18.5)`
- `WeWorkMemberType`: "2"
- `Authorization`: Bearer token from OAuth2
- `WeWorkUUID`: User UUID extracted from JWT

### Critical Field Mappings
- `SpaceID`: Must use `space.Reservable.KubeId` (a numeric string like "131834")
- `WeWorkSpaceID`: Uses `space.UUID`
- `LocationType`: Must be `2` for desk bookings
- `ApplicationType`: "WorkplaceOne"
- `PlatformType`: "iOS_APP"

### Time Handling
- Times must be in UTC format with Z suffix (e.g., "2025-06-09T23:30:00Z")
- Convert local booking times to UTC using location timezone
- Example: 08:30 Tokyo time → 2025-06-09T23:30:00Z

### Quote Response Integration
The quote response provides `CreditRatio` which must be passed to the booking request:
```go
bookingData["CreditRatio"] = quote.GrandTotal.CreditRatio
```

## Known Issues and Workarounds

- User coordinates are hardcoded in `GetAvailableSpaces` (lat: 35.68, lng: 139.69)
- No configuration file support - all config via env vars or flags
- Debug code commented out in client.go - uncomment for troubleshooting

## Development Tips

- Use `go doc` to explore package documentation
- Check `pkg/wework/types.go` for API response structures
- Browser DevTools can help understand WeWork API endpoints
- The auth flow mimics browser behavior for compatibility
- Always run `just format` in between steps. This will re-format the entire codebase and fix imports (remove unused imports, adds new imports)
