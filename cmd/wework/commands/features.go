package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewInfoCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var locationUUID string
	var city string
	var name string
	var amenitiesOnly bool
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get detailed information for a WeWork location",
		Long:  `Get detailed information, features, amenities, and instructions for a specific WeWork location.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}

			// If locationUUID is not provided, try to find it using city and name
			if locationUUID == "" {
				if city == "" || name == "" {
					return fmt.Errorf("either --location-uuid or both --city and --name must be provided")
				}

				// Get locations by city
				res, err := ww.GetLocationsByGeo(city)
				if err != nil {
					return fmt.Errorf("failed to get locations: %v", err)
				}

				// Find the location by name
				found := false
				for _, location := range res.LocationsByGeo {
					if strings.Contains(strings.ToLower(location.Name), strings.ToLower(name)) {
						locationUUID = location.UUID
						found = true
						break
					}
				}

				if !found {
					return fmt.Errorf("no location found with name containing '%s' in city '%s'", name, city)
				}
			}

			res, err := ww.GetLocationFeatures(locationUUID, amenitiesOnly)
			if err != nil {
				return fmt.Errorf("failed to get location information: %v", err)
			}

			if outputJSON {
				jsonData, err := json.MarshalIndent(res, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}
				fmt.Println(string(jsonData))
				return nil
			}

			if len(res.Locations) == 0 {
				return fmt.Errorf("no location found with UUID: %s", locationUUID)
			}

			location := res.Locations[0]
			fmt.Printf("Location: %s\n", location.Name)
			fmt.Printf("UUID: %s\n", location.UUID)
			fmt.Printf("Address: %s, %s, %s\n", location.Address.Line1, location.Address.City, location.Address.Country)
			fmt.Printf("Support Email: %s\n", location.SupportEmail)
			fmt.Printf("Phone: %s\n", location.Phone)
			
			fmt.Println("\nOperating Hours:")
			for _, hours := range location.Details.OperatingHours {
				if hours.TimeOpen != "" {
					fmt.Printf("  %s: %s - %s\n", hours.DayOfWeek, hours.TimeOpen, hours.TimeClose)
				} else {
					fmt.Printf("  %s: Closed\n", hours.DayOfWeek)
				}
			}

			fmt.Println("\nAmenities:")
			for _, amenity := range location.Amenities {
				fmt.Printf("  - %s\n", amenity.Name)
			}

			if location.MemberEntranceInstructions != "" {
				fmt.Println("\nEntrance Instructions:")
				fmt.Println(location.MemberEntranceInstructions)
			}

			if location.TourInstructions != "" {
				fmt.Println("\nTour Instructions:")
				fmt.Println(location.TourInstructions)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&locationUUID, "location-uuid", "", "UUID of the WeWork location")
	cmd.Flags().StringVar(&city, "city", "", "City name (used with --name to find location)")
	cmd.Flags().StringVar(&name, "name", "", "Location name (used with --city to find location)")
	cmd.Flags().BoolVar(&amenitiesOnly, "amenities-only", false, "Only fetch amenities information")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output raw JSON response")

	return cmd
}
