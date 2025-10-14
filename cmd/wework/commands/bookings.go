package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/pkg/spinner"
	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewBookingsCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var past bool
	var startDate, endDate string

	cmd := &cobra.Command{
		Use:   "bookings",
		Short: "List your bookings",
		Long:  `List your upcoming or past WeWork bookings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}

			var bookings []*wework.Booking
			var bookingType string

			if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
				// JSON mode: keep original flow without spinner
				if past {
					bookingType = "past"
					if startDate != "" || endDate != "" {
						var start, end time.Time
						if startDate != "" {
							start, err = time.Parse("2006-01-02", startDate)
							if err != nil {
								return fmt.Errorf("invalid start date format: %v", err)
							}
						} else {
							start = time.Now().AddDate(0, 0, -30)
						}
						if endDate != "" {
							end, err = time.Parse("2006-01-02", endDate)
							if err != nil {
								return fmt.Errorf("invalid end date format: %v", err)
							}
						} else {
							end = time.Now()
						}
						bookings, err = ww.GetPastBookingsWithDates(start, end)
						if err != nil {
							return fmt.Errorf("failed to get past bookings: %v", err)
						}
					} else {
						bookings, err = ww.GetPastBookings()
						if err != nil {
							return fmt.Errorf("failed to get past bookings: %v", err)
						}
					}
				} else {
					bookingType = "upcoming"
					bookings, err = ww.GetUpcomingBookings()
					if err != nil {
						return fmt.Errorf("failed to get upcoming bookings: %v", err)
					}
				}
			} else {
				// Text mode: use spinner for fetching bookings
				if past {
					bookingType = "past"
					if startDate != "" || endDate != "" {
						var start, end time.Time
						if startDate != "" {
							start, err = time.Parse("2006-01-02", startDate)
							if err != nil {
								return fmt.Errorf("invalid start date format: %v", err)
							}
						} else {
							start = time.Now().AddDate(0, 0, -30)
						}
						if endDate != "" {
							end, err = time.Parse("2006-01-02", endDate)
							if err != nil {
								return fmt.Errorf("invalid end date format: %v", err)
							}
						} else {
							end = time.Now()
						}

						if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
							cs.Update("Fetching past bookings…")
							r, err := ww.GetPastBookingsWithDates(start, end)
							if err != nil {
								return fmt.Errorf("failed to get past bookings: %v", err)
							}
							bookings = r
							cs.Success("Fetched past bookings")
							return nil
						}); err != nil {
							return err
						}
					} else {
						if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
							cs.Update("Fetching past bookings…")
							r, err := ww.GetPastBookings()
							if err != nil {
								return fmt.Errorf("failed to get past bookings: %v", err)
							}
							bookings = r
							cs.Success("Fetched past bookings")
							return nil
						}); err != nil {
							return err
						}
					}
				} else {
					bookingType = "upcoming"
					if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
						cs.Update("Fetching upcoming bookings…")
						r, err := ww.GetUpcomingBookings()
						if err != nil {
							return fmt.Errorf("failed to get upcoming bookings: %v", err)
						}
						bookings = r
						cs.Success("Fetched upcoming bookings")
						return nil
					}); err != nil {
						return err
					}
				}
			}

			if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
				type compactBooking struct {
					UUID         string `json:"uuid"`
					Date         string `json:"date"`
					StartTime    string `json:"startTime"`
					EndTime      string `json:"endTime"`
					LocationName string `json:"locationName"`
					LocationUUID string `json:"locationUUID"`
					Address      string `json:"address"`
					City         string `json:"city"`
					Credits      string `json:"credits"`
				}

				var compact []compactBooking
				for _, booking := range bookings {
					compact = append(compact, compactBooking{
						UUID:         booking.UUID,
						Date:         booking.StartsAt.Time.Format("2006-01-02"),
						StartTime:    booking.StartsAt.Time.Format("15:04"),
						EndTime:      booking.EndsAt.Time.Format("15:04"),
						LocationName: booking.Reservable.Location.Name,
						LocationUUID: booking.Reservable.Location.UUID,
						Address:      booking.Reservable.Location.Address.Line1,
						City:         booking.Reservable.Location.Address.City,
						Credits:      booking.CreditOrder.Price,
					})
				}

				b, err := json.MarshalIndent(compact, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}
				fmt.Println(string(b))
				return nil
			}

			if len(bookings) == 0 {
				fmt.Printf("No %s bookings found.\n", bookingType)
				return nil
			}

			fmt.Printf("%-20s%-25s%-30s%-40s%s\n", "Date", "Time", "Location", "Address", "Credits Used")
			fmt.Println(strings.Repeat("-", 145))
			for _, booking := range bookings {
				localStartsAt := booking.StartsAt.Time
				localEndsAt := booking.EndsAt.Time
				timeRange := fmt.Sprintf("%s ~ %s",
					localStartsAt.Format("15:04"),
					localEndsAt.Format("15:04 (MST)"))
				isToday := localStartsAt.Format("2006-01-02") == time.Now().Format("2006-01-02")
				dateWithDay := localStartsAt.Format("2006-01-02 Mon")
				if isToday && !past {
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

	cmd.Flags().BoolVar(&past, "past", false, "Show past bookings instead of upcoming")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date for past bookings (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date for past bookings (YYYY-MM-DD)")

	return cmd
}
