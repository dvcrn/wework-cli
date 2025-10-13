package commands

import (
	"fmt"

	"github.com/dvcrn/wework-cli/pkg/wework"
)

// resolveLocationUUID retrieves a single location UUID.
// It uses the provided locationUUID if not empty, otherwise searches based on city and name.
func resolveLocationUUID(ww *wework.WeWork, city, name, locationUUID string) (string, error) {
	if locationUUID != "" {
		return locationUUID, nil
	}
	if city == "" || name == "" {
		return "", fmt.Errorf("either --location-uuid or both --city and --name are required")
	}

	cities, err := ww.GetCities()
	if err != nil {
		return "", fmt.Errorf("failed to get cities: %w", err)
	}

	matchedCities, err := wework.FindCityByFuzzyName(city, cities)
	if err != nil {
		return "", err
	}

	var allLocations []wework.GeoLocation
	for _, matchedCity := range matchedCities {
		res, err := ww.GetLocationsByGeo(matchedCity.Name)
		if err != nil {
			return "", fmt.Errorf("failed to get locations for %s: %w", matchedCity.Name, err)
		}
		allLocations = append(allLocations, res.LocationsByGeo...)
	}

	if len(allLocations) == 0 {
		return "", fmt.Errorf("no locations found in city: %s", city)
	}

	return FindLocationByFuzzyName(name, allLocations)
}
