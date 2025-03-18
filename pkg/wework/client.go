package wework

import (
	"bytes"
	"encoding/base64"
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
	
	// Extract UUID from token and add it to headers
	userUUID := extractUUIDFromToken(token)
	if userUUID != "" {
		client.headers["WeWorkUUID"] = []string{userUUID}
	}

	return &WeWork{
		client: client,
	}
}

// Helper function to extract UUID from JWT token
func extractUUIDFromToken(token string) string {
	// Split the token to get the payload
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	
	// Decode the payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	
	// Parse the JSON
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	
	// Extract the UUID
	if uuid, ok := claims["https://wework.com/user_uuid"].(string); ok {
		return uuid
	}
	
	return ""
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

func (w *WeWork) PostBooking(date time.Time, space *Workspace) (*BookingResponse, error) {
	// First get the quote
	quote, err := w.getBookingQuote(date, space)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking quote: %v", err)
	}

	// Then create the booking
	return w.createBooking(date, space, quote)
}

func (w *WeWork) getBookingQuote(date time.Time, space *Workspace) (*QuoteResponse, error) {
	loc, err := time.LoadLocation(space.Location.TimeZone)
	if err != nil {
		return nil, err
	}

	dateInTz := date.In(loc)
	startTime := dateInTz.Format("2006-01-02") + "T00:00:00" + dateInTz.Format("-07:00")
	endTime := dateInTz.Format("2006-01-02") + "T23:59:59" + dateInTz.Format("-07:00")

	quoteURL := "https://members.wework.com/workplaceone/api/common-booking/quote"
	quoteData := map[string]interface{}{
		"SpaceType": 4,
		"ReservationID": "",
		"TriggerCalendarEvent": true,
		"Notes": map[string]interface{}{
			"locationAddress": space.Location.Address.Line1,
			"locationCity": space.Location.Address.City,
			"locationState": space.Location.Address.State,
			"locationCountry": space.Location.Address.Country,
			"locationName": space.Location.Name,
		},
		"MailData": map[string]interface{}{
			"dayFormatted": dateInTz.Format("Monday, January 2"),
			"startTimeFormatted": fmt.Sprintf("%s AM", space.OpenTime),
			"endTimeFormatted": fmt.Sprintf("%s PM", space.CloseTime),
			"floorAddress": "",
			"locationAddress": space.Location.Address.Line1,
			"creditsUsed": "0",
			"Capacity": "1",
			"TimezoneUsed": fmt.Sprintf("GMT %s", space.Location.TimezoneOffset),
			"TimezoneIana": space.Location.TimeZone,
			"startDateTime": fmt.Sprintf("%s %s", dateInTz.Format("2006-01-02"), space.OpenTime),
			"endDateTime": fmt.Sprintf("%s %s", dateInTz.Format("2006-01-02"), space.CloseTime),
			"locationName": space.Location.Name,
			"locationCity": space.Location.Address.City,
			"locationCountry": space.Location.Address.Country,
			"locationState": space.Location.Address.State,
		},
		"LocationType": 0,
		"UTCOffset": space.Location.TimezoneOffset,
		"Currency": "com.wework.credits",
		"LocationID": space.Location.UUID,
		"SpaceID": space.UUID,
		"WeWorkSpaceID": space.UUID,
		"StartTime": startTime,
		"EndTime": endTime,
	}

	quoteResp, err := w.doRequest(http.MethodPost, quoteURL, quoteData)
	if err != nil {
		return nil, err
	}
	defer quoteResp.Body.Close()

	var quote QuoteResponse
	if err := json.NewDecoder(quoteResp.Body).Decode(&quote); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %v", err)
	}

	return &quote, nil
}

func (w *WeWork) createBooking(date time.Time, space *Workspace, quote *QuoteResponse) (*BookingResponse, error) {
	loc, err := time.LoadLocation(space.Location.TimeZone)
	if err != nil {
		return nil, err
	}

	dateInTz := date.In(loc)
	startTime := dateInTz.Format("2006-01-02") + "T00:00:00" + dateInTz.Format("-07:00")
	endTime := dateInTz.Format("2006-01-02") + "T23:59:59" + dateInTz.Format("-07:00")

	if daysUntilBooking := time.Until(dateInTz); daysUntilBooking > 30*24*time.Hour {
		fmt.Println("!! Booking too far in the future, will try to book anyway, make sure you check the booking is correct !!")
		daysOver := int(daysUntilBooking/(24*time.Hour) - 30)
		adjustedDate := dateInTz.AddDate(0, 0, -(daysOver + 1))

		startTime = adjustedDate.Format("2006-01-02") + "T00:00:00" + dateInTz.Format("-07:00")
		endTime = adjustedDate.Format("2006-01-02") + "T23:59:59" + dateInTz.Format("-07:00")
	}

	bookingURL := "https://members.wework.com/workplaceone/api/common-booking/"
	bookingData := map[string]interface{}{
		"ApplicationType": "WorkplaceOne",
		"PlatformType": "iOS_APP",
		"SpaceType": 4,
		"ReservationID": "",
		"TriggerCalendarEvent": true,
		"Notes": map[string]interface{}{
			"spaceName": space.Location.Name,
			"locationAddress": space.Location.Address.Line1,
			"locationCity": space.Location.Address.City,
			"locationState": space.Location.Address.State,
			"locationCountry": space.Location.Address.Country,
			"locationName": space.Location.Address.Line1,
		},
		"MailData": map[string]interface{}{
			"dayFormatted": dateInTz.Format("Monday, January 2"),
			"startTimeFormatted": fmt.Sprintf("%s AM", space.OpenTime),
			"endTimeFormatted": fmt.Sprintf("%s PM", space.CloseTime),
			"floorAddress": "",
			"locationAddress": space.Location.Address.Line1,
			"creditsUsed": "0",
			"Capacity": "1",
			"TimezoneUsed": fmt.Sprintf("GMT %s", space.Location.TimezoneOffset),
			"TimezoneIana": space.Location.TimeZone,
			"startDateTime": fmt.Sprintf("%s %s", dateInTz.Format("2006-01-02"), space.OpenTime),
			"endDateTime": fmt.Sprintf("%s %s", dateInTz.Format("2006-01-02"), space.CloseTime),
			"locationName": space.Location.Address.Line1,
			"locationCity": space.Location.Address.City,
			"locationCountry": space.Location.Address.Country,
			"locationState": space.Location.Address.State,
		},
		"LocationType": 0,
		"UTCOffset": space.Location.TimezoneOffset,
		"LocationID": space.Location.UUID,
		"SpaceID": space.UUID,
		"WeWorkSpaceID": space.UUID,
		"StartTime": startTime,
		"EndTime": endTime,
	}

	bookingResp, err := w.doRequest(http.MethodPost, bookingURL, bookingData)
	if err != nil {
		return nil, err
	}
	defer bookingResp.Body.Close()

	var result BookingResponse
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
