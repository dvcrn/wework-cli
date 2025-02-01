package commands

import (
	"fmt"
	"strings"
	"time"

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

			if city != "" && locationUUID == "" {
				res, err := ww.GetLocationsByGeo(city)
				if err != nil {
					return fmt.Errorf("failed to get locations: %v", err)
				}

				for _, location := range res.LocationsByGeo {
					availableLocations = append(availableLocations, location.Name)
					if name == location.Name {
						targetLocationUUID = location.UUID
					}
				}
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

			// by default, assume everything is in current timezone
			tz, err := time.LoadLocation("Local")
			if err != nil {
				return fmt.Errorf("failed to load local timezone: %v", err)
			}

			for i, d := range dates {
				dates[i] = d.In(tz)
			}

			// Book for each date
			for _, bookingDate := range dates {
				fmt.Printf("Checking availability for %s\n", bookingDate)
				spaces, err := ww.GetAvailableSpaces(bookingDate, []string{targetLocationUUID})
				if err != nil {
					fmt.Printf("Error getting spaces for %s: %v\n", bookingDate, err)
					continue
				}

				if len(spaces.Response.Workspaces) == 0 {
					fmt.Printf("No spaces found for %s\n", bookingDate)
					continue
				}

				if len(spaces.Response.Workspaces) > 1 {
					fmt.Println("Found multiple spaces:")
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
				fmt.Printf("Attempting to book: %s for %s\n", space.Location.Name, bookingDate)

				bookRes, err := ww.PostBooking(bookingDate, &space)
				if err != nil {
					fmt.Printf("Booking failed: %v\n", err)
					continue
				}

				if bookRes.IsErrored {
					fmt.Printf("Booking failed: %s\n", bookRes.ErrorStatusCode)
					for _, err := range bookRes.Errors {
						fmt.Printf("  %s\n", err)
					}
					continue
				}

				fmt.Printf("Booking status: %s - %s\n",
					map[bool]string{true: "Success", false: "Failed"}[bookRes.ReservationUUID != ""],
					bookRes.BookingProcessingStatus)
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
