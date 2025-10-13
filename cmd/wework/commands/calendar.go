package commands

import (
	"encoding/json"
	"fmt"

	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewCalendarCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var calendarPath string
	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Generate calendar file",
		Long:  `Generate an ICS calendar file from your WeWork bookings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}
			cal := wework.NewWeWorkCalendar(ww)
			if err := cal.GenerateCalendar(calendarPath); err != nil {
				return fmt.Errorf("failed to generate calendar: %v", err)
			}
			if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
				b, err := json.Marshal(map[string]string{"calendarPath": calendarPath})
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}
				fmt.Println(string(b))
				return nil
			}
			fmt.Printf("Calendar generated at %s\n", calendarPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&calendarPath, "calendar-path", "wework_bookings.ics", "Output path for calendar file")
	return cmd
}
