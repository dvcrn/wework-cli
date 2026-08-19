package wework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// FavoriteLocationsResponse is the payload returned by the recent-and-favorite
// endpoint. RecentLocations is only present for some space types.
type FavoriteLocationsResponse struct {
	FavoriteLocations []FavoriteLocation `json:"FavoriteLocations"`
	RecentLocations   []FavoriteLocation `json:"RecentLocations,omitempty"`
}

// FavoriteLocation describes a single favorited (or recently used) space.
type FavoriteLocation struct {
	SpaceName                string `json:"SpaceName"`
	ID                       string `json:"Id"`
	Hmy                      int    `json:"Hmy"`
	SpaceID                  string `json:"SpaceId"`
	LocationID               string `json:"LocationId"`
	FloorID                  int    `json:"FloorId"`
	LocationName             string `json:"LocationName"`
	LocationAddress          string `json:"LocationAddress"`
	ItemImage                string `json:"ItemImage"`
	LocationType             int    `json:"LocationType"`
	SpaceType                int    `json:"SpaceType"`
	ItemType                 int    `json:"ItemType"`
	LocationAccountType      int    `json:"LocationAccountType"`
	TimeZoneIanaID           string `json:"TimeZoneIanaId"`
	AccountUUID              string `json:"AccountUUID"`
	AllowBookingInOtherZones bool   `json:"AllowBookingInOtherZones"`
	OverrideZoneBeforeHours  int    `json:"OverrideZoneBeforeHours"`
	City                     string `json:"City"`
	FloorGUID                string `json:"FloorGuid"`
}

// MarkFavoriteLocationRequest is the body sent when favoriting or unfavoriting a
// location. Set IsDeleted to true to remove an existing favorite.
type MarkFavoriteLocationRequest struct {
	ID                  int    `json:"Id,omitempty"`
	LocationID          string `json:"LocationId"`
	SpaceType           int    `json:"SpaceType"`
	IsDeleted           bool   `json:"IsDeleted"`
	LocationType        int    `json:"LocationType"`
	LocationAccountType int    `json:"LocationAccountType"`
	ReservableUUID      string `json:"ReservableUUID,omitempty"`
	SpaceID             int    `json:"SpaceId"`
	InventoryName       string `json:"InventoryName,omitempty"`
	InventoryImageURL   string `json:"InventoryImageURL,omitempty"`
	PlatformType        string `json:"PlatformType"`
	ApplicationType     string `json:"ApplicationType"`
	FloorID             int    `json:"FloorId"`
}

// GetFavoriteLocations returns the member's favorite and recent locations for the
// given space type. Valid spaceType values are 0-3; the API rejects other values
// with a 400 validation error.
func (w *WeWork) GetFavoriteLocations(spaceType int) (*FavoriteLocationsResponse, error) {
	params := url.Values{}
	params.Add("requestType", "1")
	params.Add("spaceType", strconv.Itoa(spaceType))

	apiURL := "https://members.wework.com/workplaceone/api/recent-and-favorite/v2/get-recents-and-favorite-location-data?" + params.Encode()
	resp, err := w.doRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result FavoriteLocationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode favorite locations response: %w", err)
	}

	return &result, nil
}

// MarkFavoriteLocation favorites (or, with IsDeleted set, unfavorites) a location.
// The endpoint returns an empty body on success, in which case a synthetic
// {"ok": true} result is returned.
func (w *WeWork) MarkFavoriteLocation(request MarkFavoriteLocationRequest) (map[string]any, error) {
	apiURL := "https://members.wework.com/workplaceone/api/recent-and-favorite/mark-as-favorite-location"
	resp, err := w.doRequest(http.MethodPost, apiURL, request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return map[string]any{"ok": true}, nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
