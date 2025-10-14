package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/pkg/spinner"
	"github.com/dvcrn/wework-cli/pkg/tzdate"
	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
)

func FindLocationByFuzzyName(name string, locations []wework.GeoLocation) (string, error) {
	var names []string
	for _, loc := range locations {
		names = append(names, loc.Name)
	}
	matches := fuzzy.Find(name, names)
	if len(matches) == 0 {
		return "", fmt.Errorf("no location found matching '%s'", name)
	}
	if len(matches) > 1 {
		var matchNames []string
		for _, m := range matches {
			matchNames = append(matchNames, m.Str)
		}
		return "", fmt.Errorf("multiple locations found: %s", strings.Join(matchNames, ", "))
	}
	return locations[matches[0].Index].UUID, nil
}

func NewBookCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var locationUUID, city, name, date string
	cmd := &cobra.Command{
		Use:   "book",
		Short: "Book a workspace",
		Long:  `Book a workspace at a WeWork location.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}

			if locationUUID == "" && (name == "" || city == "") {
				return fmt.Errorf("--location-uuid OR (--city + --name) is required for booking")
			}

			jsonOut, _ := cmd.Flags().GetBool("json")

			// Find target location UUID
			targetLocationUUID, err := resolveLocationUUID(ww, city, name, locationUUID)
			if err != nil {
				return err
			}

			// Parse dates
			dates := []time.Time{}
			if strings.Contains(date, "~") {
				// Date range
				parts := strings.Split(date, "~")
				if len(parts) != 2 {
					return fmt.Errorf("invalid date range format. Use YYYY-MM-DD~YYYY-MM-DD")
				}

				startDate, err := tzdate.ParseInTimezone("2006-01-02", strings.TrimSpace(parts[0]), "Local")
				if err != nil {
					return fmt.Errorf("invalid start date: %v", err)
				}

				endDate, err := tzdate.ParseInTimezone("2006-01-02", strings.TrimSpace(parts[1]), "Local")
				if err != nil {
					return fmt.Errorf("invalid end date: %v", err)
				}

				for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
					dates = append(dates, d)
				}
			} else if strings.Contains(date, ",") {
				// Comma-separated list
				for _, d := range strings.Split(date, ",") {
					parsed, err := tzdate.ParseInTimezone("2006-01-02", strings.TrimSpace(d), "Local")
					if err != nil {
						return fmt.Errorf("invalid date format: %v", err)
					}
					dates = append(dates, parsed)
				}
			} else {
				// Single date
				parsed, err := tzdate.ParseInTimezone("2006-01-02", strings.TrimSpace(date), "Local")
				if err != nil {
					return fmt.Errorf("invalid date format: %v", err)
				}
				dates = append(dates, parsed)
			}

			if jsonOut {
				// JSON mode: compute results for all dates and emit once
				type resultRow struct {
					Date          string                  `json:"date"`
					SpaceUUID     string                  `json:"spaceUUID"`
					LocationUUID  string                  `json:"locationUUID"`
					LocationName  string                  `json:"locationName"`
					BookingStatus *wework.BookingResponse `json:"booking,omitempty"`
					Error         string                  `json:"error,omitempty"`
				}
				var results []resultRow

				for _, bookingDate := range dates {
					row := resultRow{Date: bookingDate.Format("2006-01-02")}

					spaces, err := ww.GetAvailableSpaces(bookingDate, []string{targetLocationUUID})
					if err != nil {
						row.Error = fmt.Sprintf("error getting spaces: %v", err)
						results = append(results, row)
						continue
					}

					if len(spaces.Response.Workspaces) == 0 {
						row.Error = "no spaces found"
						results = append(results, row)
						continue
					}

					if len(spaces.Response.Workspaces) > 1 {
						row.Error = "multiple spaces found, please specify a specific space"
						results = append(results, row)
						continue
					}

					space := spaces.Response.Workspaces[0]
					row.SpaceUUID = space.UUID
					row.LocationUUID = space.Location.UUID
					row.LocationName = space.Location.Name

					bookRes, err := ww.PostBooking(bookingDate, &space)
					if err != nil {
						row.Error = fmt.Sprintf("booking failed: %v", err)
					} else {
						row.BookingStatus = bookRes
					}
					results = append(results, row)
				}

				b, err := json.MarshalIndent(results, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}
				fmt.Println(string(b))
			} else {
				// Text mode: drive spinner updates during network calls for each date
				for _, bookingDate := range dates {
					dateStr := bookingDate.Format("2006-01-02")
					err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
						cs.Update(fmt.Sprintf("%s: finding available spaces…", dateStr))

						spaces, err := ww.GetAvailableSpaces(bookingDate, []string{targetLocationUUID})
						if err != nil {
							return fmt.Errorf("%s: error getting spaces: %v", dateStr, err)
						}
						if len(spaces.Response.Workspaces) == 0 {
							return fmt.Errorf("%s: no spaces found", dateStr)
						}
						if len(spaces.Response.Workspaces) > 1 {
							return fmt.Errorf("%s: multiple spaces found, please specify a specific space", dateStr)
						}

						space := spaces.Response.Workspaces[0]
						cs.Update(fmt.Sprintf("%s: creating booking at %s…", dateStr, space.Location.Name))

						bookRes, err := ww.PostBooking(bookingDate, &space)
						if err != nil {
							return fmt.Errorf("%s: booking failed: %v", dateStr, err)
						}

						if bookRes.BookingStatus != "BookingSuccess" {
							errMsg := fmt.Sprintf("%s: booking failed: %s", dateStr, bookRes.BookingStatus)
							for _, e := range bookRes.Errors {
								errMsg += fmt.Sprintf("\n  %s", e)
							}
							return fmt.Errorf(errMsg)
						}

						cs.Success(fmt.Sprintf("Booking successful for %s! Reservation ID: %s", dateStr, bookRes.ReservationID))
						return nil
					})

					if err != nil {
						fmt.Printf("❌ %v\n", err)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&locationUUID, "location-uuid", "", "Location UUID for booking")
	cmd.Flags().StringVar(&city, "city", "", "City name")
	cmd.Flags().StringVar(&name, "name", "", "Space name")
	cmd.Flags().StringVar(&date, "date", time.Now().Format("2006-01-02"), "Date in YYYY-MM-DD format")

	return cmd
}
