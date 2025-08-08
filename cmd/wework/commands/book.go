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

func NewBookCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var locationUUID, city, name, date string
	cmd := &cobra.Command{
		Use:   "book",
		Short: "Book a workspace",
		Long:  `Book a workspace at a WeWork location.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}

			if locationUUID == "" && (name == "" || city == "") {
				return fmt.Errorf("--location-uuid OR (--city + --name) is required for booking")
			}

			var availableLocations []string
			var targetLocationUUID = locationUUID

			// Search for location if needed
			if city != "" && locationUUID == "" {
				// Use spinner for location search
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

			// Book for each date
			for _, bookingDate := range dates {
				// Use continuous spinner for the booking process
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
						// Need to stop spinner to show multiple options
						cs.Success("Found multiple spaces")
						fmt.Println("\nFound multiple spaces:")
						for _, space := range spaces.Response.Workspaces {
							fmt.Printf("Location: %s\n", space.Location.Name)
							fmt.Printf("Reservable ID: %s\n", space.UUID)
							fmt.Printf("Location ID: %s\n", space.Location.UUID)
							fmt.Printf("Available: %d\n", space.Seat.Available)
							fmt.Println("---")
						}
						return fmt.Errorf("please specify a specific space to book")
					}

					space := spaces.Response.Workspaces[0]

					// Submit booking
					cs.Update(fmt.Sprintf("Submitting booking for %s on %s", space.Location.Name, bookingDate.Format("2006-01-02")))
					bookRes, err := ww.PostBooking(bookingDate, &space)
					if err != nil {
						return fmt.Errorf("booking failed: %v", err)
					}

					if bookRes.BookingStatus != "BookingSuccess" {
						errMsg := fmt.Sprintf("booking failed: %s", bookRes.BookingStatus)
						for _, err := range bookRes.Errors {
							errMsg += fmt.Sprintf("\n  %s", err)
						}
						return fmt.Errorf(errMsg)
					}

					cs.Success(fmt.Sprintf("Booking successful! Reservation ID: %s", bookRes.ReservationID))
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

	cmd.Flags().StringVar(&locationUUID, "location-uuid", "", "Location UUID for booking")
	cmd.Flags().StringVar(&city, "city", "", "City name")
	cmd.Flags().StringVar(&name, "name", "", "Space name")
	cmd.Flags().StringVar(&date, "date", time.Now().Format("2006-01-02"), "Date in YYYY-MM-DD format")

	return cmd
}
