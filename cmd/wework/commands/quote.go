package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/pkg/spinner"
	"github.com/dvcrn/wework-cli/pkg/tzdate"
	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

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

			jsonOut, _ := cmd.Flags().GetBool("json")

			// Find target location UUID
			targetLocationUUID, err := resolveLocationUUID(ww, city, name, locationUUID)
			if err != nil {
				return err
			}

			if targetLocationUUID == "" {
				return fmt.Errorf("could not find any space with the name '%s'", name)
			}

			// Parse dates
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

			// Data structure for results
			type resultRow struct {
				Date         string                `json:"date"`
				SpaceUUID    string                `json:"spaceUUID"`
				LocationUUID string                `json:"locationUUID"`
				LocationName string                `json:"locationName"`
				Quote        *wework.QuoteResponse `json:"quote,omitempty"`
				Error        string                `json:"error,omitempty"`
			}
			var results []resultRow

			// Process each date
			for _, bookingDate := range dates {
				row := resultRow{Date: bookingDate.Format("2006-01-02")}

				spaces, err := ww.GetAvailableSpaces(bookingDate, []string{targetLocationUUID})
				if err != nil {
					row.Error = fmt.Sprintf("error getting spaces: %v", err)
					results = append(results, row)
					continue
				}
				if len(spaces.Response.Workspaces) == 0 {
					row.Error = "no spaces found"
					results = append(results, row)
					continue
				}
				if len(spaces.Response.Workspaces) > 1 {
					row.Error = "multiple spaces found, please be more specific"
					results = append(results, row)
					continue
				}
				space := spaces.Response.Workspaces[0]
				row.SpaceUUID = space.UUID
				row.LocationUUID = space.Location.UUID
				row.LocationName = space.Location.Name
				quote, err := ww.GetBookingQuote(bookingDate, &space)
				if err != nil {
					row.Error = fmt.Sprintf("failed to get booking quote: %v", err)
				} else {
					row.Quote = quote
				}
				results = append(results, row)
			}

			// Output results
			if jsonOut {
				b, err := json.MarshalIndent(results, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}
				fmt.Println(string(b))
			} else {
				// Human-readable output with spinners
				for _, result := range results {
					if result.Error != "" {
						if result.Error == "multiple spaces found, please be more specific" {
							fmt.Println("Found multiple spaces. Please specify a more specific name or use --location-uuid.")
						} else {
							fmt.Printf("❌ %s: %s\n", result.Date, result.Error)
						}
						continue
					}

					// Show quote with spinner
					err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
						if result.Quote == nil {
							return fmt.Errorf("no quote available")
						}
						cs.Success("✔ Done")
						return nil
					})

					if err != nil {
						fmt.Printf("❌ %v\n", err)
					} else {
						fmt.Printf("\nQuote for %s on %s:\n", result.LocationName, result.Date)
						currency := strings.Replace(result.Quote.GrandTotal.Currency, "com.wework.", "", 1)
						fmt.Printf("Quote UUID: %s\n", result.Quote.UUID)
						fmt.Printf("Total Cost: %.2f %s\n", result.Quote.GrandTotal.Amount, currency)
						if result.Quote.GrandTotal.CreditRatio > 0 {
							fmt.Printf("Credit Ratio: %.2f\n", result.Quote.GrandTotal.CreditRatio)
						}
						fmt.Println()
					}
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
