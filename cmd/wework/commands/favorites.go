package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/dvcrn/wework-cli/pkg/spinner"
	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
)

func NewFavoritesCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var spaceType int
	var showRecent bool

	cmd := &cobra.Command{
		Use:     "favorites",
		Aliases: []string{"favorite", "fav", "favs"},
		Short:   "Manage favorite locations",
		Long:    `List, add, or remove your favorite WeWork locations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFavoritesList(cmd, authenticate, spaceType, showRecent)
		},
	}

	cmd.PersistentFlags().IntVar(&spaceType, "space-type", 0, "Space type (0: Desk, 1: Office, 2: Meeting Room, 3: Event Space)")
	cmd.Flags().BoolVar(&showRecent, "recent", false, "Include recent locations in the output")

	// Subcommands
	cmd.AddCommand(
		newFavoritesListCommand(authenticate),
		newFavoritesAddCommand(authenticate),
		newFavoritesRemoveCommand(authenticate),
	)

	return cmd
}

func newFavoritesListCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var spaceType int
	var showRecent bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List favorite locations",
		Long:    `List your favorite WeWork locations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFavoritesList(cmd, authenticate, spaceType, showRecent)
		},
	}

	cmd.Flags().IntVar(&spaceType, "space-type", 0, "Space type (0: Desk, 1: Office, 2: Meeting Room, 3: Event Space)")
	cmd.Flags().BoolVar(&showRecent, "recent", false, "Include recent locations in the output")

	return cmd
}

func runFavoritesList(cmd *cobra.Command, authenticate func() (*wework.WeWork, error), spaceType int, showRecent bool) error {
	if spaceType < 0 || spaceType > 3 {
		return fmt.Errorf("invalid space-type %d: must be between 0 and 3", spaceType)
	}

	ww, err := authenticate()
	if err != nil {
		return err
	}

	jsonOut, _ := cmd.Flags().GetBool("json")
	var res *wework.FavoriteLocationsResponse

	if jsonOut {
		r, err := ww.GetFavoriteLocations(spaceType)
		if err != nil {
			return fmt.Errorf("failed to get favorite locations: %w", err)
		}
		r.FavoriteLocations = deduplicateFavoriteLocations(r.FavoriteLocations)
		r.RecentLocations = deduplicateFavoriteLocations(r.RecentLocations)
		res = r

		b, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(b))
		return nil
	}

	if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
		cs.Update(fmt.Sprintf("Fetching favorites (space type %d)…", spaceType))
		r, err := ww.GetFavoriteLocations(spaceType)
		if err != nil {
			return fmt.Errorf("failed to get favorite locations: %w", err)
		}
		r.FavoriteLocations = deduplicateFavoriteLocations(r.FavoriteLocations)
		r.RecentLocations = deduplicateFavoriteLocations(r.RecentLocations)
		res = r
		cs.Success("Favorites fetched")
		return nil
	}); err != nil {
		return err
	}

	if len(res.FavoriteLocations) == 0 && (!showRecent || len(res.RecentLocations) == 0) {
		fmt.Printf("No favorite locations found (space type %d).\n", spaceType)
		return nil
	}

	if len(res.FavoriteLocations) > 0 {
		spaceTypeName := formatSpaceTypeName(spaceType)
		fmt.Printf("Favorite Locations (%s):\n", spaceTypeName)
		fmt.Printf("%-8s%-30s%-15s%-20s%-40s%s\n", "ID", "Location", "City", "Timezone", "UUID", "Address")
		fmt.Println(strings.Repeat("-", 150))
		for _, fav := range res.FavoriteLocations {
			name := fav.LocationName
			if fav.SpaceName != "" && fav.SpaceName != fav.LocationName {
				name = fmt.Sprintf("%s (%s)", fav.LocationName, fav.SpaceName)
			}
			if len(name) > 28 {
				name = name[:28]
			}
			city := fav.City
			if len(city) > 13 {
				city = city[:13]
			}
			tz := fav.TimeZoneIanaID
			if len(tz) > 18 {
				tz = tz[:18]
			}
			addr := fav.LocationAddress
			if len(addr) > 35 {
				addr = addr[:35]
			}
			fmt.Printf("%-8d%-30s%-15s%-20s%-40s%s\n",
				fav.Hmy,
				name,
				city,
				tz,
				fav.LocationID,
				addr)
		}
	}

	if showRecent && len(res.RecentLocations) > 0 {
		if len(res.FavoriteLocations) > 0 {
			fmt.Println()
		}
		fmt.Println("Recent Locations:")
		fmt.Printf("%-8s%-30s%-15s%-20s%-40s%s\n", "ID", "Location", "City", "Timezone", "UUID", "Address")
		fmt.Println(strings.Repeat("-", 150))
		for _, rec := range res.RecentLocations {
			name := rec.LocationName
			if len(name) > 28 {
				name = name[:28]
			}
			city := rec.City
			if len(city) > 13 {
				city = city[:13]
			}
			tz := rec.TimeZoneIanaID
			if len(tz) > 18 {
				tz = tz[:18]
			}
			addr := rec.LocationAddress
			if len(addr) > 35 {
				addr = addr[:35]
			}
			fmt.Printf("%-8d%-30s%-15s%-20s%-40s%s\n",
				rec.Hmy,
				name,
				city,
				tz,
				rec.LocationID,
				addr)
		}
	}

	return nil
}

func newFavoritesAddCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var locationUUID, city, name string
	var spaceType int

	cmd := &cobra.Command{
		Use:     "add [LOCATION_UUID]",
		Aliases: []string{"mark", "set"},
		Short:   "Add a location to your favorites",
		Long:    `Add a WeWork location to your favorites list by location UUID or city and name.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				arg := args[0]
				if isUUID(arg) {
					if locationUUID == "" {
						locationUUID = arg
					}
				} else if name == "" {
					name = arg
				}
			}

			if locationUUID == "" && (name == "" || city == "") {
				return fmt.Errorf("--location-uuid OR (--city + --name) is required to add a favorite")
			}

			if spaceType < 0 || spaceType > 3 {
				return fmt.Errorf("invalid space-type %d: must be between 0 and 3", spaceType)
			}

			ww, err := authenticate()
			if err != nil {
				return err
			}

			jsonOut, _ := cmd.Flags().GetBool("json")
			var targetUUID string

			if jsonOut {
				t, err := resolveLocationUUID(ww, city, name, locationUUID)
				if err != nil {
					return err
				}
				targetUUID = t
			} else {
				if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
					cs.Update("Resolving location…")
					t, err := resolveLocationUUID(ww, city, name, locationUUID)
					if err != nil {
						return err
					}
					targetUUID = t
					cs.Success("Location resolved")
					return nil
				}); err != nil {
					return err
				}
			}

			req := wework.MarkFavoriteLocationRequest{
				LocationID:          targetUUID,
				SpaceType:           spaceType,
				IsDeleted:           false,
				LocationType:        2,
				LocationAccountType: 4,
				PlatformType:        "iOS_APP",
				ApplicationType:     "WorkplaceOne",
			}

			if jsonOut {
				res, err := ww.MarkFavoriteLocation(req)
				if err != nil {
					return fmt.Errorf("failed to add favorite location: %w", err)
				}
				b, err := json.MarshalIndent(res, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(b))
				return nil
			}

			if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
				cs.Update(fmt.Sprintf("Adding %s to favorites…", targetUUID))
				_, err := ww.MarkFavoriteLocation(req)
				if err != nil {
					return fmt.Errorf("failed to add favorite location: %w", err)
				}
				cs.Success(fmt.Sprintf("Added location %s to favorites", targetUUID))
				return nil
			}); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&locationUUID, "location-uuid", "", "Location UUID to favorite")
	cmd.Flags().StringVar(&city, "city", "", "City name")
	cmd.Flags().StringVar(&name, "name", "", "Location name")
	cmd.Flags().IntVar(&spaceType, "space-type", 0, "Space type (0: Desk, 1: Office, 2: Meeting Room, 3: Event Space)")

	return cmd
}

func newFavoritesRemoveCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var locationUUID, city, name string
	var favoriteID int
	var spaceType int

	cmd := &cobra.Command{
		Use:     "remove [LOCATION_UUID|FAVORITE_ID|NAME]",
		Aliases: []string{"rm", "delete", "unmark", "unset"},
		Short:   "Remove a location from your favorites",
		Long:    `Remove a WeWork location from your favorites list by favorite ID, location UUID, location name, or city/name.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			posArg := ""
			if len(args) > 0 {
				posArg = args[0]
			}

			if posArg == "" && favoriteID == 0 && locationUUID == "" && name == "" && city == "" {
				return fmt.Errorf("either FAVORITE_ID/LOCATION_UUID/NAME (or flags --id/--location-uuid/--name) is required")
			}

			if spaceType < 0 || spaceType > 3 {
				return fmt.Errorf("invalid space-type %d: must be between 0 and 3", spaceType)
			}

			ww, err := authenticate()
			if err != nil {
				return err
			}

			jsonOut, _ := cmd.Flags().GetBool("json")

			var resolvedIDs []int
			var resolvedName string

			if jsonOut {
				ids, rName, err := resolveRemoveFavorites(ww, posArg, locationUUID, favoriteID, city, name, spaceType)
				if err != nil {
					return err
				}
				resolvedIDs = ids
				resolvedName = rName
			} else {
				if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
					cs.Update("Resolving favorite location…")
					ids, rName, err := resolveRemoveFavorites(ww, posArg, locationUUID, favoriteID, city, name, spaceType)
					if err != nil {
						return err
					}
					resolvedIDs = ids
					resolvedName = rName
					cs.Success("Favorite identified")
					return nil
				}); err != nil {
					return err
				}
			}

			var lastRes map[string]any
			for _, id := range resolvedIDs {
				req := wework.MarkFavoriteLocationRequest{
					ID:                  id,
					SpaceType:           spaceType,
					IsDeleted:           true,
					LocationType:        2,
					LocationAccountType: 4,
					PlatformType:        "iOS_APP",
					ApplicationType:     "WorkplaceOne",
				}
				res, err := ww.MarkFavoriteLocation(req)
				if err != nil {
					return fmt.Errorf("failed to remove favorite location (ID %d): %w", id, err)
				}
				lastRes = res
			}

			if jsonOut {
				b, err := json.MarshalIndent(lastRes, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(b))
				return nil
			}

			disp := "favorite location"
			if resolvedName != "" {
				disp = resolvedName
			} else if len(resolvedIDs) == 1 {
				disp = fmt.Sprintf("favorite ID %d", resolvedIDs[0])
			}

			fmt.Printf("✓ Removed %s from favorites\n", disp)
			return nil
		},
	}

	cmd.Flags().StringVar(&locationUUID, "location-uuid", "", "Location UUID to remove from favorites")
	cmd.Flags().IntVar(&favoriteID, "id", 0, "Favorite ID to remove")
	cmd.Flags().StringVar(&city, "city", "", "City name")
	cmd.Flags().StringVar(&name, "name", "", "Location name")
	cmd.Flags().IntVar(&spaceType, "space-type", 0, "Space type (0: Desk, 1: Office, 2: Meeting Room, 3: Event Space)")

	return cmd
}

func formatSpaceTypeName(spaceType int) string {
	switch spaceType {
	case 0:
		return "Desk"
	case 1:
		return "Office"
	case 2:
		return "Meeting Room"
	case 3:
		return "Event Space"
	default:
		return fmt.Sprintf("Type %d", spaceType)
	}
}

func deduplicateFavoriteLocations(locations []wework.FavoriteLocation) []wework.FavoriteLocation {
	seen := make(map[string]bool)
	var result []wework.FavoriteLocation
	for _, loc := range locations {
		if loc.LocationID != "" && seen[loc.LocationID] {
			continue
		}
		if loc.LocationID != "" {
			seen[loc.LocationID] = true
		}
		result = append(result, loc)
	}
	return result
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, r := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if r != '-' {
				return false
			}
		} else {
			if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
				return false
			}
		}
	}
	return true
}

func resolveRemoveFavorites(ww *wework.WeWork, posArg, locationUUID string, favoriteID int, city, name string, spaceType int) ([]int, string, error) {
	if favoriteID > 0 {
		return []int{favoriteID}, fmt.Sprintf("ID %d", favoriteID), nil
	}

	target := posArg
	if target == "" && locationUUID != "" {
		target = locationUUID
	}

	// If numeric string in target, treat as ID
	if target != "" {
		if id, err := strconv.Atoi(target); err == nil && id > 0 {
			return []int{id}, fmt.Sprintf("ID %d", id), nil
		}
	}

	favs, err := ww.GetFavoriteLocations(spaceType)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch favorites: %w", err)
	}

	if len(favs.FavoriteLocations) == 0 {
		return nil, "", fmt.Errorf("you have no favorite locations (space type %d)", spaceType)
	}

	// Check if target is UUID
	if isUUID(target) {
		var ids []int
		var locName string
		for _, f := range favs.FavoriteLocations {
			if f.LocationID == target {
				ids = append(ids, f.Hmy)
				if locName == "" {
					locName = f.LocationName
				}
			}
		}
		if len(ids) > 0 {
			return ids, locName, nil
		}
		return nil, "", fmt.Errorf("location UUID %s is not in your favorites", target)
	}

	// If name not set, use target
	searchName := name
	if searchName == "" {
		searchName = target
	}

	if searchName != "" {
		// Match against favorite location names
		var names []string
		for _, f := range favs.FavoriteLocations {
			names = append(names, f.LocationName)
		}
		matches := fuzzy.Find(searchName, names)
		if len(matches) > 0 {
			matchedName := names[matches[0].Index]
			var ids []int
			for _, f := range favs.FavoriteLocations {
				if f.LocationName == matchedName {
					ids = append(ids, f.Hmy)
				}
			}
			return ids, matchedName, nil
		}
	}

	// Fallback to city + name resolution
	if city != "" && searchName != "" {
		resolvedUUID, err := resolveLocationUUID(ww, city, searchName, "")
		if err != nil {
			return nil, "", err
		}
		var ids []int
		var locName string
		for _, f := range favs.FavoriteLocations {
			if f.LocationID == resolvedUUID {
				ids = append(ids, f.Hmy)
				if locName == "" {
					locName = f.LocationName
				}
			}
		}
		if len(ids) > 0 {
			return ids, locName, nil
		}
		return nil, "", fmt.Errorf("location %s is not in your favorites", resolvedUUID)
	}

	return nil, "", fmt.Errorf("could not resolve favorite location to remove: specify ID, UUID, or location name")
}
