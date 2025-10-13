package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/pkg/tzdate"
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

			jsonOut, _ := cmd.Flags().GetBool("json")

			// Find location UUIDs and timezone
			var locationUUIDs []string
			var timezone string

			if city != "" {
				cities, err := ww.GetCities()
				if err != nil {
					return fmt.Errorf("failed to get cities: %v", err)
				}
				matchedCities, err := wework.FindCityByFuzzyName(city, cities)
				if err != nil {
					return err
				}
				var allLocations []wework.GeoLocation
				for _, matchedCity := range matchedCities {
					res, err := ww.GetLocationsByGeo(matchedCity.Name)
					if err != nil {
						return fmt.Errorf("failed to get locations for %s: %v", matchedCity.Name, err)
					}
					allLocations = append(allLocations, res.LocationsByGeo...)
				}
				if len(allLocations) == 0 {
					return fmt.Errorf("no locations found in matched cities")
				}
				for _, location := range allLocations {
					locationUUIDs = append(locationUUIDs, location.UUID)
				}
				timezone = allLocations[0].TimeZone
			} else {
				locationUUIDs = strings.Split(locationUUID, ",")
				// Get timezone from first location
				resp, err := ww.GetSpacesByUUIDs([]string{locationUUIDs[0]})
				if err != nil {
					return fmt.Errorf("failed to get location details: %w", err)
				}
				if len(resp.Response.Workspaces) == 0 {
					return fmt.Errorf("no spaces found for location UUID %s", locationUUIDs[0])
				}
				timezone = resp.Response.Workspaces[0].Location.TimeZone
			}

			if timezone == "" {
				return fmt.Errorf("could not determine timezone for desks lookup")
			}

			dateParsed, err := tzdate.ParseInTimezone("2006-01-02", date, "Local")
			if err != nil {
				return err
			}

			// Get available spaces
			resp, err := ww.GetAvailableSpaces(dateParsed, locationUUIDs)
			if err != nil {
				return fmt.Errorf("failed to get available spaces: %v", err)
			}

			// Output results
			if jsonOut {
				type row struct {
					Location     string `json:"location"`
					ReservableID string `json:"reservableId"`
					LocationID   string `json:"locationId"`
					Available    int    `json:"available"`
				}
				rows := make([]row, 0, len(resp.Response.Workspaces))
				for _, space := range resp.Response.Workspaces {
					rows = append(rows, row{
						Location:     space.Location.Name,
						ReservableID: space.UUID,
						LocationID:   space.Location.UUID,
						Available:    space.Seat.Available,
					})
				}
				b, err := json.MarshalIndent(rows, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}
				fmt.Println(string(b))
			} else {
				if len(resp.Response.Workspaces) == 0 {
					return fmt.Errorf("no spaces found, or not available for the given date")
				}

				fmt.Printf("%-30s%-40s%-40s%s\n", "Location", "Reservable ID", "Location ID", "Available")
				fmt.Println(strings.Repeat("-", 120))
				for _, space := range resp.Response.Workspaces {
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
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&locationUUID, "location-uuid", "", "Location UUID")
	cmd.Flags().StringVar(&city, "city", "", "City name")
	cmd.Flags().StringVar(&date, "date", time.Now().Format("2006-01-02"), "Date in YYYY-MM-DD format")
	return cmd
}
