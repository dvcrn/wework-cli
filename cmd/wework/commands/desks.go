package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/pkg/spinner"
	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewDesksCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var locationUUID, city, date string
	cmd := &cobra.Command{
		Use:   "desks",
		Short: "List available desks",
		Long:  `List available desks at WeWork locations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}
			if locationUUID == "" && city == "" {
				return fmt.Errorf("--location-uuid or --city is required for desks lookup")
			}

			var locationUUIDs []string
			if city != "" {
				// Use spinner for location search
				result, err := spinner.RunWithSpinner(fmt.Sprintf("Getting locations in %s", city), func() (interface{}, error) {
					res, err := ww.GetLocationsByGeo(city)
					if err != nil {
						return nil, fmt.Errorf("failed to get locations: %v", err)
					}
					var uuids []string
					for _, location := range res.LocationsByGeo {
						uuids = append(uuids, location.UUID)
					}
					return uuids, nil
				})
				if err != nil {
					return err
				}
				locationUUIDs = result.([]string)
			} else {
				locationUUIDs = strings.Split(locationUUID, ",")
			}

			dateParsed, err := time.Parse("2006-01-02", date)
			if err != nil {
				return fmt.Errorf("invalid date format: %v", err)
			}

			// Get available spaces with spinner
			result, err := spinner.RunWithSpinner(fmt.Sprintf("Getting available desks for %s", dateParsed.Format("2006-01-02")), func() (interface{}, error) {
				return ww.GetAvailableSpaces(dateParsed, locationUUIDs)
			})
			if err != nil {
				return fmt.Errorf("failed to get available spaces: %v", err)
			}
			response := result.(*wework.SharedWorkspaceResponse)

			if len(response.Response.Workspaces) == 0 {
				return fmt.Errorf("no spaces found, or not available for the given date")
			}

			fmt.Printf("%-30s%-40s%-40s%s\n", "Location", "Reservable ID", "Location ID", "Available")
			fmt.Println(strings.Repeat("-", 120))
			for _, space := range response.Response.Workspaces {
				name := space.Location.Name
				if len(name) > 28 {
					name = name[:28]
				}
				fmt.Printf("%-30s%-40s%-40s%d\n",
					name,
					space.UUID,
					space.Location.UUID,
					space.Seat.Available)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&locationUUID, "location-uuid", "", "Location UUID")
	cmd.Flags().StringVar(&city, "city", "", "City name")
	cmd.Flags().StringVar(&date, "date", time.Now().Format("2006-01-02"), "Date in YYYY-MM-DD format")
	return cmd
}
