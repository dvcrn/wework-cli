package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/cmd/wework/commands"
	"github.com/dvcrn/wework-cli/pkg/wework"

	"github.com/spf13/cobra"
)

var (
	username         string
	password         string
	locationUUID     string
	city             string
	name             string
	date             string
	calendarPath     string
	includeBootstrap bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "wework",
		Short: "WeWork CLI tool",
		Long:  `A command line interface for WeWork workspace booking and management.`,
	}

	rootCmd.PersistentFlags().StringVar(&username, "username", os.Getenv("WEWORK_USERNAME"), "WeWork username")
	rootCmd.PersistentFlags().StringVar(&password, "password", os.Getenv("WEWORK_PASSWORD"), "WeWork password")


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

	// Calendar command
	calendarCmd := &cobra.Command{
		Use:   "calendar",
		Short: "Generate calendar file",
		Long:  `Generate an ICS calendar file from your WeWork bookings.`,
		RunE:  runCalendar,
	}
	calendarCmd.Flags().StringVar(&calendarPath, "calendar-path", "wework_bookings.ics", "Output path for calendar file")

	meCmd := &cobra.Command{
		Use:   "me",
		Short: "Get your profile information",
		Long:  `Get your profile information from WeWork.`,
		RunE:  runMe,
	}
	meCmd.Flags().BoolVar(&includeBootstrap, "include-bootstrap", false, "Include bootstrap data in profile information")

	rootCmd.AddCommand(
		locationsCmd,
		desksCmd,
		bookingsCmd,
		commands.NewBookCommand(authenticate),
		calendarCmd,
		meCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func authenticate() (*wework.WeWork, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required. Set WEWORK_USERNAME and WEWORK_PASSWORD environment variables or use --username and --password flags")
	}

	weworkAuth, err := wework.NewWeWorkAuth(username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create WeWork auth: %v", err)
	}

	loginResult, _, err := weworkAuth.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	return wework.NewWeWork(loginResult.AccessToken, loginResult.A0token), nil
}

func runMe(cmd *cobra.Command, args []string) error {
	ww, err := authenticate()
	if err != nil {
		return err
	}

	userResponse, err := ww.GetUserProfile()
	if err != nil {
		return fmt.Errorf("failed to get user profile: %v", err)
	}

	fmt.Printf("User Profile:\n")
	fmt.Printf("  UUID: %s\n", userResponse.UUID)
	fmt.Printf("  Name: %s\n", userResponse.Name)
	fmt.Printf("  Email: %s\n", userResponse.Email)
	fmt.Printf("  Phone: %s\n", userResponse.Phone)
	fmt.Printf("  Language: %s\n", userResponse.LanguagePreference)
	fmt.Printf("  Is WeWork: %v\n", userResponse.IsWework)
	fmt.Printf("  Is Admin: %v\n", userResponse.IsAdmin)
	fmt.Printf("  Active: %v\n", userResponse.Active)

	fmt.Printf("\nHome Location:\n")
	fmt.Printf("  Name: %s\n", userResponse.HomeLocation.Name)
	fmt.Printf("  Address: %s, %s, %s %s\n",
		userResponse.HomeLocation.Address.Line1,
		userResponse.HomeLocation.Address.City,
		userResponse.HomeLocation.Address.State,
		userResponse.HomeLocation.Address.Zip)
	fmt.Printf("  Timezone: %s\n", userResponse.HomeLocation.TimeZone)

	fmt.Printf("\nCompanies:\n")
	for _, company := range userResponse.Companies {
		fmt.Printf("  - %s (UUID: %s)\n", company.Name, company.UUID)
		if company.PreferredMembershipNullable != nil {
			fmt.Printf("    Membership: %s\n", company.PreferredMembershipNullable.MembershipType)
		}
	}

	if !includeBootstrap {
		return nil
	}

	bootstrap, err := ww.GetBootstrap()
	if err != nil {
		return fmt.Errorf("failed to get bootstrap: %v", err)
	}

	fmt.Printf("\nBootstrap Data:\n")
	fmt.Printf("  Menu Security Data:\n")
	fmt.Printf("    Password Change Enforcing: %v\n", bootstrap.MenuSecurityData.IsPasswordChangeEnforcing)
	fmt.Printf("    Admin Role: %s\n", bootstrap.MenuSecurityData.AdminRole)
	fmt.Printf("    Cache Response: %v\n", bootstrap.MenuSecurityData.Cacheresponse)

	fmt.Printf("\n  Page Sentry Data:\n")
	for _, item := range bootstrap.PageSentryData.AllowedItems {
		fmt.Printf("    - ID: %d, URL: %s\n", item.Identifier, item.URLFragment)
	}

	fmt.Printf("\n  User Data:\n")
	fmt.Printf("    Email: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserEmail)
	fmt.Printf("    Name: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserName)
	fmt.Printf("    Phone: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserPhone)
	fmt.Printf("    Membership UUID: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkMembershipUUID)
	fmt.Printf("    Product UUID: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkProductUUID)
	fmt.Printf("    Preferred Membership UUID: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkPreferredMembershipUUID)
	fmt.Printf("    User UUID: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserUUID)
	fmt.Printf("    Company UUID: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkCompanyUUID)
	fmt.Printf("    Chargable Account UUID: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkChargableAccountUUID)
	fmt.Printf("    Company Migration Status: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkCompanyMigrationStatus)
	fmt.Printf("    Membership Type: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkMembershipType)
	fmt.Printf("    Membership Name: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkMembershipName)
	fmt.Printf("    Avatar: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserAvatar)
	fmt.Printf("    Language Preference: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserLanguagePreference)
	fmt.Printf("    Home Location UUID: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserHomeLocationUUID)
	fmt.Printf("    Home Location Name: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserHomeLocationName)
	fmt.Printf("    Home Location City: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserHomeLocationCity)
	fmt.Printf("    Home Location Latitude: %f\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserHomeLocationLatitude)
	fmt.Printf("    Home Location Longitude: %f\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserHomeLocationLongitude)
	fmt.Printf("    Home Location Migrated: %v\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserHomeLocationMigrated)
	fmt.Printf("    Preferred Currency: %s\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserPreferredCurrency)
	fmt.Printf("    No Active Memberships: %v\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.NoActiveMemberships)
	fmt.Printf("    Is Kube Migrated Account: %v\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.IsKubeMigratedAccount)
	fmt.Printf("    Theme Preference: %d\n", bootstrap.WeworkUserProfileData.ProfileData.WeWorkUserData.WeWorkUserThemePreference)

	fmt.Printf("\n  Company Data:\n")
	for _, company := range bootstrap.WeworkUserProfileData.ProfileData.WeWorkCompanyList {
		fmt.Printf("    - Company: %s\n", company.CompanyName)
		fmt.Printf("      UUID: %s\n", company.CompanyUUID)
		fmt.Printf("      License Type: %d\n", company.CompanyLicenseType)
		fmt.Printf("      Preferred Membership UUID: %s\n", company.PreferredMembershipUUID)
		fmt.Printf("      Preferred Membership Name: %s\n", company.PreferredMembershipName)
		fmt.Printf("      Preferred Membership Type: %s\n", company.PreferredMembershipType)
		fmt.Printf("      Preferred Membership Product UUID: %s\n", company.PreferredMembershipProductUUID)
		fmt.Printf("      Is Migrated To KUBE: %v\n", company.IsMigratedToKUBE)
		fmt.Printf("      Migration Status: %s\n", company.CompanyMigrationStatus)
		fmt.Printf("      KUBE Company UUID: %s\n", company.KUBECompanyUUID)
	}

	fmt.Printf("\n  Memberships:\n")
	for _, membership := range bootstrap.WeworkUserProfileData.ProfileData.WeWorkMembershipsList {
		fmt.Printf("    - UUID: %s\n", membership.UUID)
		fmt.Printf("      Account UUID: %s\n", membership.AccountUUID)
		fmt.Printf("      Type: %s\n", membership.MembershipType)
		fmt.Printf("      User UUID: %s\n", membership.UserUUID)
		fmt.Printf("      Product Name: %s\n", membership.ProductName)
		fmt.Printf("      Product UUID: %s\n", membership.ProductUUID)
		fmt.Printf("      Activated On: %s\n", membership.ActivatedOn)
		fmt.Printf("      Started On: %s\n", membership.StartedOn)
		fmt.Printf("      Is Migrated: %v\n", membership.IsMigrated)
		fmt.Printf("      Priority Order: %d\n", membership.PriorityOrder)
	}

	fmt.Printf("\n  Profile Data:\n")
	fmt.Printf("    User Onboarding Status: %v\n", bootstrap.WeworkUserProfileData.ProfileData.UserOnboardingStatus)
	fmt.Printf("    Debug Mode Enabled: %v\n", bootstrap.WeworkUserProfileData.ProfileData.DebugModeEnabled)
	fmt.Printf("    Is User Workplace Admin: %v\n", bootstrap.WeworkUserProfileData.ProfileData.IsUserWorkplaceAdmin)
	fmt.Printf("    Account Manager Link: %s\n", bootstrap.WeworkUserProfileData.ProfileData.AccountManagerLink)

	fmt.Printf("\n  Experience Status:\n")
	fmt.Printf("    Workplace Experience: %v\n", bootstrap.WorkplaceExperienceStatus)
	fmt.Printf("    Vast Experience: %v\n", bootstrap.VastExperienceStatus)

	fmt.Printf("\n  Global Settings:\n")
	fmt.Printf("    Allow Affiliate Bookings: %v\n", bootstrap.GlobalSettings.AllowAffiliateBookingsInMemberWeb)
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

func runCalendar(cmd *cobra.Command, args []string) error {
	ww, err := authenticate()
	if err != nil {
		return err
	}

	cal := wework.NewWeWorkCalendar(ww)
	if err := cal.GenerateCalendar(calendarPath); err != nil {
		return fmt.Errorf("failed to generate calendar: %v", err)
	}

	fmt.Printf("Calendar generated at %s\n", calendarPath)

	return nil
}
