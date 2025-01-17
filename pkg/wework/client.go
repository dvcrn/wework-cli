package wework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WeWork struct {
	client  *http.Client
	headers http.Header
}

func NewWeWork(authorization, weworkAuth string) *WeWork {
	headers := http.Header{
		"Referer":          []string{"https://members.wework.com/workplaceone/content2/bookings/desks"},
		"Authorization":    []string{"Bearer " + authorization},
		"Accept":           []string{"application/json, text/plain, */*"},
		"Content-Type":     []string{"application/json"},
		"WeWorkAuth":       []string{"Bearer " + weworkAuth},
		"IsKube":           []string{"true"},
		"Request-Source":   []string{"MemberWeb/WorkplaceOne/Prod"},
		"WeWorkMemberType": []string{"2"},
	}

	return &WeWork{
		client:  &http.Client{},
		headers: headers,
	}
}

func (w *WeWork) doRequest(method, url string, data interface{}) (*http.Response, error) {
	var body []byte
	var err error

	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %v", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header = w.headers.Clone()

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		var errorResp struct {
			ResponseStatus struct {
				Type    string `json:"type"`
				Message string `json:"message"`
				Title   string `json:"title"`
			} `json:"responseStatus"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if errorResp.ResponseStatus.Type == "error" {
				return nil, fmt.Errorf("API error: %s (%s)", errorResp.ResponseStatus.Message, errorResp.ResponseStatus.Title)
			}
		}

		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func (w *WeWork) GetLocationsByGeo(city string) (*LocationsByGeoResponse, error) {
	params := url.Values{}
	params.Add("isAuthenticated", "true")
	params.Add("city", city)
	params.Add("isOnDemandUser", "false")
	params.Add("isWeb", "true")

	url := "https://members.wework.com/workplaceone/api/wework-yardi/ondemand/get-locations-by-geo?" + params.Encode()
	resp, err := w.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result LocationsByGeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

func (w *WeWork) GetAvailableSpaces(t time.Time, locationUUIDs []string) (*SharedWorkspaceResponse, error) {

	params := url.Values{}
	params.Add("locationUUIDs", strings.Join(locationUUIDs, ","))
	params.Add("closestCity", "")
	params.Add("userLatitude", "35.6953443")
	params.Add("userLongitude", "139.7564755")
	params.Add("boundnwLat", "")
	params.Add("boundnwLng", "")
	params.Add("boundseLat", "")
	params.Add("boundseLng", "")
	params.Add("type", "0")
	params.Add("offset", "0")
	params.Add("limit", "50")
	params.Add("roomTypeFilter", "")
	params.Add("date", t.Format("2006-01-02"))
	params.Add("duration", "30")
	params.Add("locationOffset", "+09:00")
	params.Add("isWeb", "true")
	params.Add("capacity", "0")
	params.Add("endDate", "")

	url := "https://members.wework.com/workplaceone/api/spaces/get-spaces?" + params.Encode()
	resp, err := w.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	reader := bytes.NewReader(buf.Bytes())

	//b := bytes.Buffer{}
	//b.ReadFrom(reader)
	//fmt.Println(b.String())
	//
	//reader.Seek(0, 0)
	var result SharedWorkspaceResponse
	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

func (w *WeWork) GetUpcomingBookings() ([]*Booking, error) {
	url := "https://members.wework.com/workplaceone/api/ext-booking/get-wework-upcoming-booking-data"
	resp, err := w.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// bodyBuf := new(bytes.Buffer)
	// bodyBuf.ReadFrom(resp.Body)

	// reader := bytes.NewReader(bodyBuf.Bytes())

	// buf := new(bytes.Buffer)
	// buf.ReadFrom(reader)
	// fmt.Println(buf.String())

	// reader.Seek(0, 0)
	var result []*Booking
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	for _, booking := range result {
		w.adjustBookingTimezone(booking)
	}

	return result, nil
}

func (w *WeWork) adjustBookingTimezone(booking *Booking) {
	loc, err := time.LoadLocation(booking.Reservable.Location.TimeZone)
	if err != nil {
		return
	}

	booking.StartsAt = booking.StartsAt.In(loc)
	booking.EndsAt = booking.EndsAt.In(loc)
}

func (w *WeWork) GetPastBookings() ([]*Booking, error) {
	url := "https://members.wework.com/workplaceone/api/ext-booking/get-wework-past-booking-data"
	resp, err := w.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []*Booking
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	for _, booking := range result {
		w.adjustBookingTimezone(booking)
	}

	return result, nil
}

func (w *WeWork) MxgPollQuote(date time.Time, space *Workspace) (*BookingQuote, error) {
	loc, err := time.LoadLocation(space.Location.TimeZone)
	if err != nil {
		return nil, err
	}

	dateInTz := date.In(loc)
	startTime := dateInTz.Format("2006-01-02") + "T00:00:00" + dateInTz.Format("-07:00")
	endTime := dateInTz.Format("2006-01-02") + "T23:59:00" + dateInTz.Format("-07:00")

	url := "https://members.wework.com/workplaceone/api/ext-booking/mxg-poll-quote?APIen=0"
	data := map[string]interface{}{
		"reservableId":         space.UUID,
		"type":                 4,
		"creditsUsed":          0,
		"currency":             "com.wework.credits",
		"TriggerCalenderEvent": true,
		"mailData": map[string]interface{}{
			"dayFormatted":       dateInTz.Format("Monday, January 2"),
			"startTimeFormatted": "08:30:00 AM",
			"endTimeFormatted":   "20:00 PM",
			"floorAddress":       "",
			"locationAddress":    space.Location.Address.Line1,
			"creditsUsed":        "0",
			"Capacity":           "1",
			"TimezoneUsed":       fmt.Sprintf("GMT %s", space.Location.TimeZone),
			"TimezoneIana":       space.Location.TimeZone,
			"TimezoneWin":        space.Location.TimeZone,
			"startDateTime":      fmt.Sprintf("%s 08:30", date.Format("2006-01-02")),
			"endDateTime":        fmt.Sprintf("%s 20:00", date.Format("2006-01-02")),
			"locationName":       space.Location.Name,
			"locationCity":       space.Location.Address.City,
			"locationCountry":    space.Location.Address.Country,
			"locationState":      space.Location.Address.State,
		},
		"coworkingPropertyId": 0,
		"locationId":          space.Location.UUID,
		"Notes": map[string]interface{}{
			"locationAddress": space.Location.Address.Line1,
			"locationCity":    space.Location.Address.City,
			"locationState":   space.Location.Address.State,
			"locationCountry": space.Location.Address.Country,
			"locationName":    space.Location.Name,
		},
		"isUpdateBooking": false,
		"reservationId":   "",
		"startTime":       startTime,
		"endTime":         endTime,
	}

	resp, err := w.doRequest(http.MethodPost, url, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result BookingQuote
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

func (w *WeWork) PostBooking(date time.Time, space *Workspace) (*BookSpaceResponse, error) {
	loc, err := time.LoadLocation(space.Location.TimeZone)
	if err != nil {
		return nil, err
	}

	dateInTz := date.In(loc)

	// start/end time should be in format 2025-01-06T00:00:00+UTC OFFSET
	startTime := dateInTz.Format("2006-01-02") + "T00:00:00" + dateInTz.Format("-07:00")
	endTime := dateInTz.Format("2006-01-02") + "T23:59:00" + dateInTz.Format("-07:00")

	quote, err := w.MxgPollQuote(date, space)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %v", err)
	}

	url := "https://members.wework.com/workplaceone/api/ext-booking/post-booking?APIen=0"
	data := map[string]interface{}{
		"reservableId":         space.UUID,
		"type":                 4,
		"creditsUsed":          0,
		"orderId":              quote.UUID,
		"ApplicationType":      "WorkplaceOne",
		"PlatformType":         "WEB",
		"TriggerCalenderEvent": true,
		"mailData": map[string]interface{}{
			"dayFormatted":       dateInTz.Format("Monday, January 2"),
			"startTimeFormatted": "08:30:00 AM",
			"endTimeFormatted":   "20:00 PM",
			"floorAddress":       "",
			"locationAddress":    space.Location.Address.Line1,
			"creditsUsed":        "0",
			"Capacity":           "1",
			"TimezoneUsed":       fmt.Sprintf("GMT %s", space.Location.TimeZone),
			"TimezoneIana":       space.Location.TimeZone,
			"TimezoneWin":        space.Location.TimeZone,
			"startDateTime":      fmt.Sprintf("%s 08:30", date.Format("2006-01-02")),
			"endDateTime":        fmt.Sprintf("%s 20:00", date.Format("2006-01-02")),
			"locationName":       space.Location.Name,
			"locationCity":       space.Location.Address.City,
			"locationCountry":    space.Location.Address.Country,
			"locationState":      space.Location.Address.State,
		},
		"coworkingPropertyId": 0,
		"applicationType":     "WorkplaceOne",
		"platformType":        "WEB",
		"locationId":          space.Location.UUID,
		"Notes": map[string]interface{}{
			"locationAddress": space.Location.Address.Line1,
			"locationCity":    space.Location.Address.City,
			"locationState":   space.Location.Address.State,
			"locationCountry": space.Location.Address.Country,
			"locationName":    space.Location.Name,
		},
		"isUpdateBooking": false,
		"reservationId":   "",
		"startTime":       startTime,
		"endTime":         endTime,
	}

	resp, err := w.doRequest(http.MethodPost, url, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result BookSpaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}
