package wework

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type CancelBookingRequest struct {
	Method  string         `json:"method"`
	URL     string         `json:"url"`
	Headers http.Header    `json:"headers"`
	Body    map[string]any `json:"body"`
}

type CancelBookingResponse struct {
	Raw json.RawMessage `json:"raw"`
}

func (w *WeWork) BuildCancelBookingRequest(bookingUUID string) (*CancelBookingRequest, error) {
	booking, err := w.findUpcomingBooking(bookingUUID)
	if err != nil {
		return nil, err
	}

	return buildCancelBookingRequest(booking)
}

func (w *WeWork) CancelBooking(bookingUUID string) (*CancelBookingResponse, error) {
	req, err := w.BuildCancelBookingRequest(bookingUUID)
	if err != nil {
		return nil, err
	}

	return w.cancelBooking(req)
}

func (w *WeWork) findUpcomingBooking(bookingUUID string) (*Booking, error) {
	bookings, err := w.GetUpcomingBookings()
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming bookings: %w", err)
	}

	for _, booking := range bookings {
		if booking.UUID == bookingUUID {
			return booking, nil
		}
	}

	return nil, fmt.Errorf("upcoming booking %s not found", bookingUUID)
}

func buildCancelBookingRequest(booking *Booking) (*CancelBookingRequest, error) {
	if booking == nil {
		return nil, fmt.Errorf("booking cannot be nil")
	}
	if booking.Reservable == nil {
		return nil, fmt.Errorf("booking %s is missing reservable details", booking.UUID)
	}
	if booking.Reservable.Location == nil {
		return nil, fmt.Errorf("booking %s is missing location details", booking.UUID)
	}

	bookingID := booking.KubeBookingExternalReference
	if bookingID == "" {
		bookingID = booking.UUID
	}

	startTime := formatCancelTime(booking.StartsAt.Time)
	endTime := formatCancelTime(booking.EndsAt.Time)
	location := booking.Reservable.Location
	reservableID := booking.Reservable.UUID

	cancelData := map[string]any{
		"bookingId":           bookingID,
		"bookingLocationType": location.AccountType,
		"creditsUsed":         creditsUsed(booking),
		"startTime":           startTime,
		"endTime":             endTime,
		"locationId":          location.UUID,
		"reservableId":        reservableID,
		"isBookingApprovalOn": booking.IsBookingApprovalOn,
		"bookingType":         4,
		"spaceId":             reservableID,
		"cancellationNote":    "",
		"mailParams": map[string]any{
			"workspaceType":      1,
			"dayFormatted":       formatDayWithOrdinal(booking.StartsAt.Time),
			"startTimeFormatted": startTime,
			"endTimeFormatted":   endTime,
			"floorAddress":       "",
			"locationAddress":    formatCancelLocationAddress(location.Address),
			"locationCountry":    location.Address.Country,
		},
		"reservationId": bookingID,
	}

	params := url.Values{}
	params.Add("isOnDemand", "false")
	params.Add("platFormType", "1")

	return &CancelBookingRequest{
		Method: http.MethodPost,
		URL:    "https://members.wework.com/workplaceone/api/common-booking/cancel?" + params.Encode(),
		Headers: http.Header{
			"Request-Source": []string{"MemberWeb/WorkplaceOne/Prod"},
			"Referer":        []string{"https://members.wework.com/workplaceone/content2/your-bookings"},
			"fe-pg":          []string{"/workplaceone/content2/your-bookings"},
		},
		Body: cancelData,
	}, nil
}

func (w *WeWork) cancelBooking(cancelReq *CancelBookingRequest) (*CancelBookingResponse, error) {
	resp, err := w.doRequestWithHeaders(cancelReq.Method, cancelReq.URL, cancelReq.Body, cancelReq.Headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode cancel response: %w", err)
	}

	return &CancelBookingResponse{Raw: raw}, nil
}

func formatCancelTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000")
}

func creditsUsed(booking *Booking) float64 {
	if booking.CreditOrder == nil {
		return 0
	}
	credits, err := strconv.ParseFloat(booking.CreditOrder.Price, 64)
	if err != nil {
		return 0
	}
	return credits
}

func formatCancelLocationAddress(address Address) string {
	parts := []string{address.Line1, address.Line2}
	var nonEmpty []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			nonEmpty = append(nonEmpty, part)
		}
	}
	return strings.Join(nonEmpty, " ")
}

func formatDayWithOrdinal(t time.Time) string {
	return fmt.Sprintf("%s, %s %s", t.Weekday(), t.Month(), ordinalDay(t.Day()))
}

func ordinalDay(day int) string {
	if day%100 >= 11 && day%100 <= 13 {
		return fmt.Sprintf("%dth", day)
	}

	switch day % 10 {
	case 1:
		return fmt.Sprintf("%dst", day)
	case 2:
		return fmt.Sprintf("%dnd", day)
	case 3:
		return fmt.Sprintf("%drd", day)
	default:
		return fmt.Sprintf("%dth", day)
	}
}
