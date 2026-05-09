package dto

type TopupValidateAccountRequest struct {
	SKUID        int64  `json:"sku_id"`
	MobileNumber string `json:"mobile_number"`
}

type TopupValidateAccountResponse struct {
	SKUID        int64  `json:"sku_id"`
	MobileNumber string `json:"mobile_number"`
	Valid        bool   `json:"valid"`
	Operator     string `json:"operator"`
	Message      string `json:"message"`
}

type FlightSearchRequest struct {
	From  string
	To    string
	Date  string
	Adult int
}

type FlightSearchResponse struct {
	RouteCode string        `json:"route_code"`
	From      string        `json:"from"`
	To        string        `json:"to"`
	Date      string        `json:"date"`
	Adult     int           `json:"adult"`
	Offers    []FlightOffer `json:"offers"`
}

type FlightOffer struct {
	OfferToken       string   `json:"offer_token"`
	SKUID            int64    `json:"sku_id"`
	FlightNo         string   `json:"flight_no"`
	Carrier          string   `json:"carrier"`
	DepartureTime    string   `json:"departure_time"`
	ArrivalTime      string   `json:"arrival_time"`
	Price            MoneyDTO `json:"price"`
	Seats            int      `json:"seats"`
	OfferType        string   `json:"offer_type"`
	AvailabilityType string   `json:"availability_type"`
	BookingMode      string   `json:"booking_mode"`
}
