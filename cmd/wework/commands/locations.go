package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewLocationsCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var city string
	cmd := &cobra.Command{
		Use:   "locations",
		Short: "List WeWork locations",
		Long:  `List available WeWork locations in a city.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}
			res, err := ww.GetLocationsByGeo(city)
			if err != nil {
				return fmt.Errorf("failed to get locations: %v", err)
			}
			if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
				b, err := json.MarshalIndent(res.LocationsByGeo, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}
				fmt.Println(string(b))
				return nil
			}
			fmt.Printf("%-30s%-40s%-15s%s\n", "Location", "UUID", "Latitude", "Longitude")
			fmt.Println(strings.Repeat("-", 95))
			for _, location := range res.LocationsByGeo {
				name := location.Name
				if len(name) > 28 {
					name = name[:28]
				}
				fmt.Printf("%-30s%-40s%-15.6f%f\n",
					name,
					location.UUID,
					location.Latitude,
					location.Longitude)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&city, "city", "", "City name")
	cmd.MarkFlagRequired("city")
	return cmd
}
