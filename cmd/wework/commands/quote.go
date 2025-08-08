package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/pkg/spinner"
	"github.com/dvcrn/wework-cli/pkg/tzdate"
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
				startDate, err := tzdate.ParseInTimezone("2006-01-02", strings.TrimSpace(parts[0]), "Local")
				if err != nil {
					return fmt.Errorf("invalid start date: %v", err)
				}
				endDate, err := tzdate.ParseInTimezone("2006-01-02", strings.TrimSpace(parts[1]), "Local")
				if err != nil {
					return fmt.Errorf("invalid end date: %v", err)
				}
				for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
					dates = append(dates, d)
				}
			} else if strings.Contains(date, ",") {
				// Comma-separated list
				for _, d := range strings.Split(date, ",") {
					parsed, err := tzdate.ParseInTimezone("2006-01-02", strings.TrimSpace(d), "Local")
					if err != nil {
						return fmt.Errorf("invalid date format: %v", err)
					}
					dates = append(dates, parsed)
				}
			} else {
				// Single date
				parsed, err := tzdate.ParseInTimezone("2006-01-02", strings.TrimSpace(date), "Local")
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
				var finalQuote *wework.QuoteResponse
				var finalSpace wework.Workspace

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

					// Handle multiple spaces without printing inside the spinner
					if len(spaces.Response.Workspaces) > 1 {
						// This case will be handled outside the spinner
						return fmt.Errorf("multiple spaces found, please be more specific")
					}

					finalSpace = spaces.Response.Workspaces[0]

					// Get quote
					cs.Update(fmt.Sprintf("Getting quote for %s on %s", finalSpace.Location.Name, bookingDate.Format("2006-01-02")))
					quote, err := ww.GetBookingQuote(bookingDate, &finalSpace)
					if err != nil {
						return fmt.Errorf("failed to get booking quote: %w", err)
					}
					finalQuote = quote

					cs.Success("✔ Done")
					return nil
				})

				if err != nil {
					// Handle the specific case of multiple spaces cleanly
					if err.Error() == "multiple spaces found, please be more specific" {
						fmt.Println("Found multiple spaces. Please specify a more specific name or use --location-uuid.")
					} else {
						fmt.Printf("❌ %v\n", err)
					}
					continue
				}

				if finalQuote != nil {
					fmt.Printf("\nQuote for %s on %s:\n", finalSpace.Location.Name, bookingDate.Format("2006-01-02"))
					currency := strings.Replace(finalQuote.GrandTotal.Currency, "com.wework.", "", 1)
					fmt.Printf("Quote UUID: %s\n", finalQuote.UUID)
					fmt.Printf("Total Cost: %.2f %s\n", finalQuote.GrandTotal.Amount, currency)
					if finalQuote.GrandTotal.CreditRatio > 0 {
						fmt.Printf("Credit Ratio: %.2f\n", finalQuote.GrandTotal.CreditRatio)
					}
					fmt.Println()
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
