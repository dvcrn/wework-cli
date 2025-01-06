package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/pkg/auth"
	"github.com/dvcrn/wework-cli/pkg/wework"

	"github.com/spf13/cobra"
)

var (
	username     string
	password     string
	locationUUID string
	city         string
	name         string
	date         string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "wework",
		Short: "WeWork CLI tool",
		Long:  `A command line interface for WeWork workspace booking and management.`,
	}

	rootCmd.PersistentFlags().StringVar(&username, "username", os.Getenv("WEWORK_USERNAME"), "WeWork username")
	rootCmd.PersistentFlags().StringVar(&password, "password", os.Getenv("WEWORK_PASSWORD"), "WeWork password")

	//Book command
	bookCmd := &cobra.Command{
		Use:   "book",
		Short: "Book a workspace",
		Long:  `Book a workspace at a WeWork location.`,
		RunE:  runBook,
	}
	bookCmd.Flags().StringVar(&locationUUID, "location-uuid", "", "Location UUID for booking")
	bookCmd.Flags().StringVar(&city, "city", "", "City name")
	bookCmd.Flags().StringVar(&name, "name", "", "Space name")
	bookCmd.Flags().StringVar(&date, "date", time.Now().Format("2006-01-02"), "Date in YYYY-MM-DD format")

	// Locations command
	locationsCmd := &cobra.Command{
		Use:   "locations",
		Short: "List WeWork locations",
		Long:  `List available WeWork locations in a city.`,
		RunE:  runLocations,
	}
	locationsCmd.Flags().StringVar(&city, "city", "", "City name")
	locationsCmd.MarkFlagRequired("city")

	// Desks command
	desksCmd := &cobra.Command{
		Use:   "desks",
		Short: "List available desks",
		Long:  `List available desks at WeWork locations.`,
		RunE:  runDesks,
	}
	desksCmd.Flags().StringVar(&locationUUID, "location-uuid", "", "Location UUID")
	desksCmd.Flags().StringVar(&city, "city", "", "City name")
	desksCmd.Flags().StringVar(&date, "date", time.Now().Format("2006-01-02"), "Date in YYYY-MM-DD format")

	// Bookings command
	bookingsCmd := &cobra.Command{
		Use:   "bookings",
		Short: "List your bookings",
		Long:  `List your upcoming WeWork bookings.`,
		RunE:  runBookings,
	}

	rootCmd.AddCommand(locationsCmd, desksCmd, bookingsCmd, bookCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func authenticate() (*wework.WeWork, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required. Set WEWORK_USERNAME and WEWORK_PASSWORD environment variables or use --username and --password flags")
	}

	weworkAuth, err := auth.NewWeWorkAuth(username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create WeWork auth: %v", err)
	}

	result, err := weworkAuth.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	return wework.NewWeWork(result.Token, result.IDToken), nil
}

func runBook(cmd *cobra.Command, args []string) error {
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

		// TODO: for debug. remove me.
		func(v interface{}) {
			j, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			buf := bytes.NewBuffer(j)
			fmt.Printf("%v\n", buf.String())
		}(spaces)

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

		// convert bookingDate to timezone of space

		space := spaces.Response.Workspaces[0]
		fmt.Printf("Attempting to book: %s for %s\n", space.Location.Name, bookingDate)

		bookRes, err := ww.PostBooking(bookingDate, &space)
		if err != nil {
			fmt.Printf("Booking failed: %v\n", err)
			continue
		}
		fmt.Printf("Booking status: %s - %s\n",
			map[bool]string{true: "Success", false: "Failed"}[bookRes.ReservationUUID != ""],
			bookRes.BookingProcessingStatus)
	}

	return nil
}

func runLocations(cmd *cobra.Command, args []string) error {
	ww, err := authenticate()
	if err != nil {
		return err
	}

	res, err := ww.GetLocationsByGeo(city)
	if err != nil {
		return fmt.Errorf("failed to get locations: %v", err)
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
}

func runDesks(cmd *cobra.Command, args []string) error {
	ww, err := authenticate()
	if err != nil {
		return err
	}

	if locationUUID == "" && city == "" {
		return fmt.Errorf("--location-uuid or --city is required for desks lookup")
	}

	var locationUUIDs []string
	if city != "" {
		res, err := ww.GetLocationsByGeo(city)
		if err != nil {
			return fmt.Errorf("failed to get locations: %v", err)
		}

		// TODO: for debug. remove me.
		func(v interface{}) {
			j, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			buf := bytes.NewBuffer(j)
			fmt.Printf("%v\n", buf.String())
		}(res)

		for _, location := range res.LocationsByGeo {
			locationUUIDs = append(locationUUIDs, location.UUID)
		}
	} else {
		locationUUIDs = strings.Split(locationUUID, ",")
	}

	dateParsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format: %v", err)
	}

	response, err := ww.GetAvailableSpaces(dateParsed, locationUUIDs)
	if err != nil {
		return fmt.Errorf("failed to get available spaces: %v", err)
	}

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
}

func runBookings(cmd *cobra.Command, args []string) error {
	ww, err := authenticate()
	if err != nil {
		return err
	}

	bookings, err := ww.GetUpcomingBookings()
	if err != nil {
		return fmt.Errorf("failed to get upcoming bookings: %v", err)
	}

	if len(bookings) == 0 {
		fmt.Println("No upcoming bookings found.")
		return nil
	}

	fmt.Printf("%-20s%-25s%-30s%-40s%s\n", "Date", "Time", "Location", "Address", "Credits Used")
	fmt.Println(strings.Repeat("-", 145))

	for _, booking := range bookings {
		localStartsAt := booking.StartsAt
		localEndsAt := booking.EndsAt
		timeRange := fmt.Sprintf("%s ~ %s",
			localStartsAt.Format("15:04"),
			localEndsAt.Format("15:04 (MST)"))

		isToday := localStartsAt.Format("2006-01-02") == time.Now().Format("2006-01-02")
		dateWithDay := localStartsAt.Format("2006-01-02 Mon")
		if isToday {
			dateWithDay += " *"
		}

		name := booking.Reservable.Location.Name
		if len(name) > 28 {
			name = name[:28]
		}

		address := booking.Reservable.Location.Address.Line1
		if len(address) > 38 {
			address = address[:38]
		}

		fmt.Printf("%-20s%-25s%-30s%-40s%s\n",
			dateWithDay,
			timeRange,
			name,
			address,
			booking.CreditOrder.Price)
	}

	return nil
}
