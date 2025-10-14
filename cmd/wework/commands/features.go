package commands

import (
	"encoding/json"
	"fmt"

	"github.com/dvcrn/wework-cli/pkg/spinner"
	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewInfoCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var locationUUID string
	var city string
	var name string
	var amenitiesOnly bool

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get detailed information for a WeWork location",
		Long:  `Get detailed information, features, amenities, and instructions for a specific WeWork location.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}

			var res *wework.LocationFeaturesResponse
			if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
				// JSON: no spinner
				if locationUUID == "" {
					if city == "" || name == "" {
						return fmt.Errorf("either --location-uuid or both --city and --name must be provided")
					}
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
						r, err := ww.GetLocationsByGeo(matchedCity.Name)
						if err != nil {
							return fmt.Errorf("failed to get locations for %s: %v", matchedCity.Name, err)
						}
						allLocations = append(allLocations, r.LocationsByGeo...)
					}
					var err2 error
					locationUUID, err2 = FindLocationByFuzzyName(name, allLocations)
					if err2 != nil {
						return err2
					}
				}
				r, err := ww.GetLocationFeatures(locationUUID, amenitiesOnly)
				if err != nil {
					return fmt.Errorf("failed to get location information: %v", err)
				}
				res = r
			} else {
				// Text: use spinner across resolution and fetch
				if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
					if locationUUID == "" {
						if city == "" || name == "" {
							return fmt.Errorf("either --location-uuid or both --city and --name must be provided")
						}
						cs.Update("Fetching cities…")
						cities, err := ww.GetCities()
						if err != nil {
							return fmt.Errorf("failed to get cities: %v", err)
						}
						cs.Update("Matching city…")
						matchedCities, err := wework.FindCityByFuzzyName(city, cities)
						if err != nil {
							return err
						}
						var allLocations []wework.GeoLocation
						for _, matchedCity := range matchedCities {
							cs.Update(fmt.Sprintf("Fetching locations for %s…", matchedCity.Name))
							r, err := ww.GetLocationsByGeo(matchedCity.Name)
							if err != nil {
								return fmt.Errorf("failed to get locations for %s: %v", matchedCity.Name, err)
							}
							allLocations = append(allLocations, r.LocationsByGeo...)
						}
						var err2 error
						locationUUID, err2 = FindLocationByFuzzyName(name, allLocations)
						if err2 != nil {
							return err2
						}
					}
					cs.Update("Fetching location features…")
					r, err := ww.GetLocationFeatures(locationUUID, amenitiesOnly)
					if err != nil {
						return fmt.Errorf("failed to get location information: %v", err)
					}
					res = r
					cs.Success("Information retrieved")
					return nil
				}); err != nil {
					return err
				}
			}

			if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
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

	return cmd
}
