package commands

import (
	"encoding/json"
	"fmt"

	"github.com/dvcrn/wework-cli/pkg/spinner"
	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewCancelCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel BOOKING_UUID",
		Short: "Cancel an upcoming booking",
		Long:  `Cancel an upcoming WeWork booking by the booking UUID shown by the bookings command.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ww, err := authenticate()
			if err != nil {
				return err
			}

			bookingUUID := args[0]
			jsonOut, _ := cmd.Flags().GetBool("json")

			var res *wework.CancelBookingResponse
			if jsonOut {
				res, err = ww.CancelBooking(bookingUUID)
				if err != nil {
					return fmt.Errorf("failed to cancel booking %s: %w", bookingUUID, err)
				}

				b, err := json.MarshalIndent(res.Raw, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(b))
				return nil
			}

			if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
				cs.Update(fmt.Sprintf("Canceling booking %s…", bookingUUID))
				res, err = ww.CancelBooking(bookingUUID)
				if err != nil {
					return fmt.Errorf("failed to cancel booking %s: %w", bookingUUID, err)
				}
				cs.Success(fmt.Sprintf("Canceled booking %s", bookingUUID))
				return nil
			}); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
