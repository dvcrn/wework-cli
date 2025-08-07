package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/pkg/spinner"
	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

// NewQuoteCommand creates the 'quote' command.
// Its structure mirrors NewBookCommand but calls GetBookingQuote instead of PostBooking.
func NewQuoteCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var locationUUID, city, name, date string

	cmd := &cobra.Command{
		Use:   "quote",
		Short: "Get a booking quote for a workspace",
		Long:  `Get a booking quote for a workspace at a WeWork location without creating a booking. This is useful for testing availability and pricing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}

			if locationUUID == "" && (name == "" || city == "") {
				return fmt.Errorf("--location-uuid OR (--city + --name) is required for quoting")
			}

			var availableLocations []string
			var targetLocationUUID = locationUUID

			// Search for location if needed, same as book.go
			if city != "" && locationUUID == "" {
				result, err := spinner.RunWithSpinner(fmt.Sprintf("Searching for locations in %s", city), func() (interface{}, error) {
					res, err := ww.GetLocationsByGeo(city)
					if err != nil {
						return nil, fmt.Errorf("failed to get locations: %v", err)
					}

					var foundUUID string
					for _, location := range res.LocationsByGeo {
						availableLocations = append(availableLocations, location.Name)
						if name == location.Name {
							foundUUID = location.UUID
						}
					}
					return foundUUID, nil
				})

				if err != nil {
					return err
				}

				targetLocationUUID = result.(string)
			}

			if targetLocationUUID == "" {
				return fmt.Errorf("could not find any space with the name '%s'. Available locations for city %s are: %s",
					name, city, strings.Join(availableLocations, ", "))
			}

			// Parse dates, same as book.go
			dates := []time.Time{}
			if strings.Contains(date, "~") {
				// Date range
				parts := strings.Split(date, "~")
				if len(parts) != 2 {
					return fmt.Errorf("invalid date range format. Use YYYY-MM-DD~YYYY-MM-DD")
				}
				startDate, err := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
				if err != nil {
					return fmt.Errorf("invalid start date: %v", err)
				}
				endDate, err := time.Parse("2006-01-02", strings.TrimSpace(parts[1]))
				if err != nil {
					return fmt.Errorf("invalid end date: %v", err)
				}
				for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
					dates = append(dates, d)
				}
			} else if strings.Contains(date, ",") {
				// Comma-separated list
				for _, d := range strings.Split(date, ",") {
					parsed, err := time.Parse("2006-01-02", strings.TrimSpace(d))
					if err != nil {
						return fmt.Errorf("invalid date format: %v", err)
					}
					dates = append(dates, parsed)
				}
			} else {
				// Single date
				parsed, err := time.Parse("2006-01-02", strings.TrimSpace(date))
				if err != nil {
					return fmt.Errorf("invalid date format: %v", err)
				}
				dates = append(dates, parsed)
			}

			tz, err := time.LoadLocation("Local")
			if err != nil {
				return fmt.Errorf("failed to load local timezone: %v", err)
			}
			for i, d := range dates {
				dates[i] = d.In(tz)
			}

			// Get quote for each date
			for _, bookingDate := range dates {
				err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
					// Check availability
					cs.Update(fmt.Sprintf("Checking availability for %s", bookingDate.Format("2006-01-02")))
					spaces, err := ww.GetAvailableSpaces(bookingDate, []string{targetLocationUUID})
					if err != nil {
						return fmt.Errorf("error getting spaces for %s: %v", bookingDate, err)
					}
					if len(spaces.Response.Workspaces) == 0 {
						return fmt.Errorf("no spaces found for %s", bookingDate)
					}

					if len(spaces.Response.Workspaces) > 1 {
						cs.Success("Found multiple spaces")
						fmt.Println("\nFound multiple spaces:")
						for _, space := range spaces.Response.Workspaces {
							fmt.Printf("Location: %s\n", space.Location.Name)
							if space.Reservable != nil {
								fmt.Printf("Reservable KubeID: %s\n", space.Reservable.KubeId)
							}
							fmt.Printf("Space UUID: %s\n", space.UUID)
							fmt.Printf("Available: %d\n", space.Seat.Available)
							fmt.Println("---")
						}
						return fmt.Errorf("please specify a more specific space to quote, or use the UUIDs provided")
					}

					space := spaces.Response.Workspaces[0]

					// Get quote
					cs.Update(fmt.Sprintf("Getting quote for %s on %s", space.Location.Name, bookingDate.Format("2006-01-02")))
					quote, err := ww.GetBookingQuote(bookingDate, &space)
					if err != nil {
						return fmt.Errorf("failed to get booking quote: %w", err)
					}

					// Print the quote response as nicely formatted JSON
					quoteJSON, err := json.MarshalIndent(quote, "", "  ")
					if err != nil {
						return fmt.Errorf("failed to format quote response into JSON: %w", err)
					}

					cs.Success(fmt.Sprintf("Quote for %s on %s:", space.Location.Name, bookingDate.Format("2006-01-02")))
					fmt.Println(string(quoteJSON))

					return nil
				})

				if err != nil {
					fmt.Printf("❌ %v\n", err)
					continue
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&locationUUID, "location-uuid", "", "Location UUID for quoting")
	cmd.Flags().StringVar(&city, "city", "", "City name")
	cmd.Flags().StringVar(&name, "name", "", "Space name")
	cmd.Flags().StringVar(&date, "date", time.Now().Format("2006-01-02"), "Date in YYYY-MM-DD format (can be a single date, a comma-separated list, or a range like YYYY-MM-DD~YYYY-MM-DD)")

	return cmd
}
