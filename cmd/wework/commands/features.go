package commands

import (
	"encoding/json"
	"fmt"

	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewFeaturesCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var locationUUID string
	var amenitiesOnly bool
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "features",
		Short: "Get features and instructions for a WeWork location",
		Long:  `Get detailed features, amenities, and instructions for a specific WeWork location.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}

			res, err := ww.GetLocationFeatures(locationUUID, amenitiesOnly)
			if err != nil {
				return fmt.Errorf("failed to get location features: %v", err)
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
	cmd.Flags().BoolVar(&amenitiesOnly, "amenities-only", false, "Only fetch amenities information")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output raw JSON response")
	cmd.MarkFlagRequired("location-uuid")

	return cmd
}
