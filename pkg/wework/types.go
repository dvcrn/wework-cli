package wework

import "time"

type SharedWorkspaceResponse struct {
	Limit    int `json:"limit"`
	Offset   int `json:"offset"`
	Response struct {
		Workspaces []Workspace `json:"workspaces"`
	} `json:"getSharedWorkspaces"`
}

type LocationsByGeoResponse struct {
	LocationsByGeo []GeoLocation `json:"locationsByGeo"`
}

type BookSpaceResponse struct {
	BookingProcessingStatus string   `json:"bookingProcessingStatus"`
	Errors                  []string `json:"errors"`
	IsErrored               bool     `json:"isErrorred"`
	ReservationUUID         string   `json:"reservationUUID"`
}

type GeoLocation struct {
	UUID                 string  `json:"uuid"`
	Name                 string  `json:"name"`
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	Address              Address `json:"address"`
	TimeZone             string  `json:"timeZone"`
	Distance             float64 `json:"distance"`
	BrandName            string  `json:"brandName"`
	HasThirdPartyDisplay bool    `json:"hasThirdPartyDisplay"`
	Image                string  `json:"image"`
	IsMigrated           bool    `json:"isMigrated"`
}

type UpcomingBookingsResponse struct {
	Bookings []Booking `json:"bookings"`
}

type Booking struct {
	UUID                         string           `json:"uuid"`
	StartsAt                     time.Time        `json:"startsAt"`
	EndsAt                       time.Time        `json:"endsAt"`
	TimeZone                     string           `json:"timezone"`
	CreditOrder                  *CreditOrder     `json:"creditOrder"`
	Reservable                   *SharedWorkspace `json:"reservable"`
	IsAttendee                   bool             `json:"isAttendee"`
	ModificationDeadline         time.Time        `json:"modificationDeadline"`
	Order                        Order            `json:"order"`
	IsMultidayBooking            bool             `json:"isMultidayBooking"`
	KubeSameDayCancelPolicy      bool             `json:"kubeSameDayCancelPolicy"`
	IsFromKube                   bool             `json:"isFromKube"`
	KubeBookingExternalReference string           `json:"kubeBookingExternalReference"`
	CwmBookingReferenceID        int              `json:"cwmBookingReferenceId"`
	IsFromCwm                    bool             `json:"isFromCwm"`
	IsBookingConfirmationPending bool             `json:"isBookingConfirmationPending"`
	IsBookingApprovalOn          bool             `json:"IsBookingApprovalOn"`
	SameDayCancelPolicy          bool             `json:"sameDayCancelPolicy"`
	// KubeCreatedOnDate            *time.Time      `json:"kubeCreatedOnDate,omitempty"`
	// KubeModifiedOnDate           *time.Time      `json:"kubeModifiedOnDate,omitempty"`
	// KubeStartDate                *time.Time      `json:"kubeStartDate,omitempty"`
}

type CreditOrder struct {
	Price string `json:"price"`
}

type SharedWorkspace struct {
	UUID       string                   `json:"uuid"`
	Capacity   int                      `json:"capacity"`
	TypeName   string                   `json:"__typename"`
	Location   *SharedWorkspaceLocation `json:"location"`
	ImageURL   string                   `json:"imageUrl"`
	CwmSpaceID int                      `json:"cwmSpaceId"`
}

type SharedWorkspaceLocation struct {
	KubePropertyID       int     `json:"kubePropertyID"`
	CwmPropertyID        int     `json:"cwmPropertyID"`
	UUID                 string  `json:"uuid"`
	Name                 string  `json:"name"`
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	Address              Address `json:"address"`
	TimeZone             string  `json:"timeZone"`
	Distance             float64 `json:"distance"`
	HasThirdPartyDisplay bool    `json:"hasThirdPartyDisplay"`
	IsMigrated           bool    `json:"isMigrated"`
}

type Order struct {
	PaymentProfileUUID string `json:"paymentProfileUuid"`
	SubTotal           Amount `json:"subTotal"`
	Adjustments        []any  `json:"adjustments"`
	Taxes              []any  `json:"taxes"`
	GrandTotal         Amount `json:"grandTotal"`
}

type Amount struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Address struct {
	Line1   string `json:"line1"`
	Line2   string `json:"line2"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
	Zip     string `json:"zip"`
}

type Workspace struct {
	UUID                 string            `json:"uuid"`
	InventoryUUID        string            `json:"inventoryUuid"`
	ImageURL             string            `json:"imageURL"`
	HeaderImageURL       string            `json:"headerImageUrl"`
	Capacity             int               `json:"capacity"`
	Credits              int               `json:"credits"`
	Location             Location          `json:"location"`
	OpenTime             string            `json:"openTime"`
	CloseTime            string            `json:"closeTime"`
	CancellationPolicy   string            `json:"cancellationPolicy"`
	OperatingHours       []*OperatingHours `json:"operatingHours"`
	ProductPrice         *ProductPrice     `json:"productPrice"`
	Seat                 Seat              `json:"seat"`
	SeatsAvailable       int               `json:"seatsAvailable"`
	Order                any               `json:"order"`
	IsVASTCoworking      bool              `json:"isVASTCoworking"`
	IsAffiliateCoworking bool              `json:"isAffiliateCoworking"`
	IsFranchiseCoworking bool              `json:"isFranchiseCoworking"`
	IsHybridSpace        bool              `json:"isHybridSpace"`
}

type Location struct {
	Description                string      `json:"description"`
	SupportEmail               string      `json:"supportEmail"`
	PhoneNormalized            string      `json:"phoneNormalized"`
	Currency                   string      `json:"currency"`
	PrimaryTeamMember          TeamMember  `json:"primaryTeamMember"`
	Amenities                  []Amenity   `json:"amenities"`
	Details                    Details     `json:"details"`
	TransitInfo                TransitInfo `json:"transitInfo"`
	MemberEntranceInstructions string      `json:"memberEntranceInstructions"`
	ParkingInstructions        string      `json:"parkingInstructions"`
	TimezoneOffset             string      `json:"timezoneOffset"`
	TimeZoneIdentifier         string      `json:"timeZoneIdentifier"`
	TimeZoneWinID              string      `json:"timeZoneWinId"`
	UUID                       string      `json:"uuid"`
	Name                       string      `json:"name"`
	Latitude                   float64     `json:"latitude"`
	Longitude                  float64     `json:"longitude"`
	Address                    Address     `json:"address"`
	TimeZone                   string      `json:"timeZone"`
	Distance                   float64     `json:"distance"`
	HasThirdPartyDisplay       bool        `json:"hasThirdPartyDisplay"`
	IsMigrated                 bool        `json:"isMigrated"`
}

type TeamMember struct {
	Name          string `json:"name"`
	BusinessTitle string `json:"businessTitle"`
	ImageURL      string `json:"imageUrl"`
}

type Amenity struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Highlight bool   `json:"highlight"`
}

type Details struct {
	HasExtendedHours bool `json:"hasExtendedHours"`
}

type TransitInfo struct {
	Bike    string `json:"bike"`
	Bus     string `json:"bus"`
	Ferry   string `json:"ferry"`
	Freeway string `json:"freeway"`
	Metro   string `json:"metro"`
	Parking string `json:"parking"`
}

type OperatingHours struct {
	DayOfWeek int    `json:"dayOfWeek"`
	Day       string `json:"day"`
	Open      string `json:"open"`
	Close     string `json:"close"`
	IsClosed  bool   `json:"isClosed"`
}

type ProductPrice struct {
	UUID                 string                 `json:"uuid"`
	ProductUUID          string                 `json:"productUuid"`
	Price                Price                  `json:"price"`
	RateUnit             int                    `json:"rateUnit"`
	HalfHourCreditPrices []*HalfHourCreditPrice `json:"halfHourCreditPrices"`
}

type Price struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type HalfHourCreditPrice struct {
	Offset int     `json:"offset"`
	Amount float64 `json:"amount"`
}

type Seat struct {
	Total     int `json:"total"`
	Available int `json:"available"`
}

type BookingQuote struct {
	UUID          string      `json:"uuid"`
	QuoteStatus   int         `json:"quoteStatus"`
	StatusDetails []string    `json:"statusDetails"`
	GrandTotal    *Price      `json:"grandTotal"`
	SubTotal      *Price      `json:"subTotal"`
	Taxes         []*Price    `json:"taxes"`
	LineItems     []*LineItem `json:"lineItems"`
	Adjustments   []*Price    `json:"adjustments"`
}

type LineItem struct{}
