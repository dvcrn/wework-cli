package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewBookingsCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bookings",
		Short: "List your bookings",
		Long:  `List your upcoming WeWork bookings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
	return cmd
}
