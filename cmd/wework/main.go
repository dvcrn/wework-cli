package main

import (
	"fmt"
	"os"

	"github.com/dvcrn/wework-cli/cmd/wework/commands"
	"github.com/dvcrn/wework-cli/pkg/spinner"
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

	rootCmd.AddCommand(
		commands.NewLocationsCommand(authenticate),
		commands.NewDesksCommand(authenticate),
		commands.NewBookingsCommand(authenticate),
		commands.NewBookCommand(authenticate),
		commands.NewCalendarCommand(authenticate),
		commands.NewMeCommand(authenticate),
		commands.NewInfoCommand(authenticate),
		commands.NewQuoteCommand(authenticate),
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

	// Run authentication with spinner
	result, err := spinner.RunWithSpinner("Authenticating with WeWork", func() (interface{}, error) {
		weworkAuth, err := wework.NewWeWorkAuth(username, password)
		if err != nil {
			return nil, fmt.Errorf("failed to create WeWork auth: %v", err)
		}

		loginResult, _, err := weworkAuth.Authenticate()
		if err != nil {
			return nil, fmt.Errorf("authentication failed: %v", err)
		}

		// Use the same token for both authorization headers
		return wework.NewWeWork(loginResult.A0token), nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*wework.WeWork), nil
}
