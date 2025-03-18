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
	client *BaseClient
}

func NewWeWork(token string) *WeWork {
	client, err := NewBaseClient()
	if err != nil {
		panic(err)
	}

	// Use the same token for both headers
	client.headers["Authorization"] = []string{"Bearer " + token}
	client.headers["WeWorkAuth"] = []string{"Bearer " + token}
	
	// Update headers to match the working request
	client.headers["Request-Source"] = []string{"MemberWeb/WorkplaceOne/Prod"}
	client.headers["User-Agent"] = []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Safari/605.1.15"}
	client.headers["fe-pg"] = []string{"/workplaceone/content2/dashboard"}
	client.headers["Sec-Fetch-Site"] = []string{"same-origin"}
	client.headers["Sec-Fetch-Mode"] = []string{"cors"}
	client.headers["Sec-Fetch-Dest"] = []string{"empty"}

	return &WeWork{
		client: client,
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

		// buf := new(bytes.Buffer)
		// buf.ReadFrom(resp.Body)
		// resp.Body = io.NopCloser(buf)

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
	url := "https://members.wework.com/workplaceone/api/common-booking/upcoming-bookings"
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
	var result UpcomingBookingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	for _, booking := range result.Bookings {
		w.adjustBookingTimezone(booking)
	}

	return result.Bookings, nil
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

func (w *WeWork) GetBootstrap() (*AppBootstrapResponse, error) {
	url := "https://members.wework.com/workplaceone/api/app-bootstrap/bootstrap"

	data := map[string]interface{}{
		"InvalidateCache": false,
		"platform":        1,
		"FeatureFlags": map[string]interface{}{
			"WeGateMemberWebFlags": []string{
				"WG_WEWORK_W_HOMEPAGE_PRINTING",
				"WG_WEWORK_W_MEMWEB_ANNOUNCEMENTS_FROM_CONTENTFUL",
				"WG_WEWORK_W_MEMWEB_BUILDING_GUIDE_UPCOMING_BOOKINGS",
				"WG_WEWORK_W_MEMWEB_ENTERPRISE",
				"WG_WEWORK_W_MEMWEB_EVENTS",
				"WG_WEWORK_W_MEMWEB_SUPPORT_HELP_FAQ",
				"WG_WEWORK_W_MEMWEB_TOP_BANNER_ALL_ACCESS",
				"WG_WEWORK_W_MEMWEB_WEWORK_BRANDING",
				"WG_WEWORK_W_ROOMS_MEDALLIA_SURVEY",
				"WG_WEWORK_W_MEMWEB_WEB_THIRD_PARTY_SPACES",
				"WG_WEWORK_W_MEMWEB_GUEST_POLICY_ENFORCEMENT",
				"WG_WEWORK_W_MEMWEB_PRINT_DRIVER_UPDATE_ROLLOUT",
				"WG_WEWORK_W_MEMWEB_BUILDING_GUIDE_ORGANON_MODULES",
			},
			"WeGateiOSFlags":     []string{},
			"WeGateAndroidFlags": []string{},
		},
		"PermissionRequest": map[string]interface{}{
			"MENAflags": []string{
				"mena_module_building_guide_categories",
				"mena_module_account_manager",
				"mena_module_daily_desks",
				"mena_module_print_hub",
				"mena_module_events",
			},
		},
		"AppVersion":         nil,
		"CurrentAccountUUID": "",
	}

	resp, err := w.doRequest(http.MethodPost, url, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result AppBootstrapResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

func (w *WeWork) GetUserProfile() (*UserProfileResponse, error) {
	url := "https://members.wework.com/workplaceone/api/wework-yardi/user/get-user-profile"
	resp, err := w.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result UserProfileResponse
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
	startTime := dateInTz.Format("2006-01-02") + "T00:00:00" + dateInTz.Format("-07:00")
	endTime := dateInTz.Format("2006-01-02") + "T23:59:59" + dateInTz.Format("-07:00")

	// First request - mxg-poll-quote
	quoteURL := "https://members.wework.com/workplaceone/api/ext-booking/mxg-poll-quote?APIen=0"
	quoteData := map[string]interface{}{
		"reservableId":         space.UUID,
		"type":                 4,
		"creditsUsed":          1,
		"currency":             "com.wework.credits",
		"TriggerCalenderEvent": true,
		"mailData": map[string]interface{}{
			"dayFormatted":       dateInTz.Format("Monday, January 2"),
			"startTimeFormatted": fmt.Sprintf("%s AM", space.OpenTime),
			"endTimeFormatted":   fmt.Sprintf("%s AM", space.CloseTime),
			"floorAddress":       "",
			"locationAddress":    space.Location.Address.Line1,
			"creditsUsed":        "1",
			"Capacity":           "1",
			"TimezoneUsed":       fmt.Sprintf("GMT %s", space.Location.TimeZone),
			"TimezoneIana":       space.Location.TimeZone,
			"startDateTime":      fmt.Sprintf("%s %s", dateInTz.Format("2006-01-02"), space.OpenTime),
			"endDateTime":        fmt.Sprintf("%s %s", dateInTz.Format("2006-01-02"), space.CloseTime),
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

	quoteResp, err := w.doRequest(http.MethodPost, quoteURL, quoteData)
	if err != nil {
		return nil, err
	}
	defer quoteResp.Body.Close()

	var quote BookingQuote
	if err := json.NewDecoder(quoteResp.Body).Decode(&quote); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %v", err)
	}

	if daysUntilBooking := time.Until(dateInTz); daysUntilBooking > 30*24*time.Hour {
		fmt.Println("!! Booking too far in the future, will try to book anyway, make sure you check the booking is correct !!")
		daysOver := int(daysUntilBooking/(24*time.Hour) - 30)
		adjustedDate := dateInTz.AddDate(0, 0, -(daysOver + 1))

		startTime = adjustedDate.Format("2006-01-02") + "T00:00:00" + dateInTz.Format("-07:00")
		endTime = adjustedDate.Format("2006-01-02") + "T23:59:59" + dateInTz.Format("-07:00")
	}

	// Second request - post-booking
	bookingURL := "https://members.wework.com/workplaceone/api/ext-booking/post-booking?APIen=0"
	bookingData := map[string]interface{}{
		"reservableId":         space.UUID,
		"type":                 4,
		"creditsUsed":          quote.GrandTotal.Amount,
		"orderId":              quote.UUID,
		"ApplicationType":      "WorkplaceOne",
		"PlatformType":         "iOS_APP",
		"TriggerCalenderEvent": true,
		"mailData": map[string]interface{}{
			"dayFormatted":       dateInTz.Format("Monday, January 2"),
			"startTimeFormatted": fmt.Sprintf("%s AM", space.OpenTime),
			"endTimeFormatted":   fmt.Sprintf("%s AM", space.CloseTime),
			"floorAddress":       "",
			"locationAddress":    space.Location.Address.Line1,
			"creditsUsed":        "1",
			"Capacity":           space.Capacity,
			"TimezoneUsed":       space.Location.TimezoneOffset,
			"TimezoneIana":       space.Location.TimeZone,
			// "startDateTime":      "2025-02-26 08:30",
			// "endDateTime":        "2025-02-26 20:00",
			"startDateTime":   fmt.Sprintf("%s %s", dateInTz.Format("2006-01-02"), space.OpenTime),
			"endDateTime":     fmt.Sprintf("%s %s", dateInTz.Format("2006-01-02"), space.CloseTime),
			"locationName":    space.Location.Address.Line1,
			"locationCity":    space.Location.Address.City,
			"locationCountry": space.Location.Address.Country,
			"locationState":   space.Location.Address.State,
		},
		"coworkingPropertyId": 0,
		"applicationType":     "WorkplaceOne",
		"platformType":        "iOS_APP",
		"locationId":          space.Location.UUID,
		"Notes": map[string]interface{}{
			"spaceName":       space.Location.Name,
			"locationAddress": space.Location.Address.Line1,
			"locationCity":    space.Location.Address.City,
			"locationState":   space.Location.Address.State,
			"locationCountry": space.Location.Address.Country,
			"locationName":    space.Location.Address.Line1,
		},
		"isUpdateBooking": false,
		"reservationId":   "",
		// "startTime":       "2025-02-26T00:00:00+09:00",
		// "endTime":         "2025-02-26T23:59:59+09:00",
		"startTime": startTime,
		"endTime":   endTime,
	}

	bookingResp, err := w.doRequest(http.MethodPost, bookingURL, bookingData)
	if err != nil {
		return nil, err
	}
	defer bookingResp.Body.Close()

	var result BookSpaceResponse
	if err := json.NewDecoder(bookingResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode booking response: %v", err)
	}

	return &result, nil
}

func (w *WeWork) GetCityDetails() ([]*CityDetailsResponse, error) {
	url := "https://members.wework.com/workplaceone/api/wework-yardi/location/get-city-details"
	resp, err := w.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []*CityDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result, nil
}
