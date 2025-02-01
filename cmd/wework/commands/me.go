package commands

import (
	"fmt"

	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewMeCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var includeBootstrap bool
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Get your profile information",
		Long:  `Get your profile information from WeWork.`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
	cmd.Flags().BoolVar(&includeBootstrap, "include-bootstrap", false, "Include bootstrap data in profile information")
	return cmd
}
