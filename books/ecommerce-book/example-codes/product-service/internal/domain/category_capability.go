package domain

// ProductModelType describes how a category represents sellable products.
type ProductModelType string

const (
	ProductModelSingleSKU     ProductModelType = "SINGLE_SKU"
	ProductModelAccountBased  ProductModelType = "ACCOUNT_BASED"
	ProductModelResourceBased ProductModelType = "RESOURCE_BASED"
	ProductModelRealtimeOffer ProductModelType = "REALTIME_OFFER"
)

// ResourceType describes the stable resource behind a sellable product.
type ResourceType string

const (
	ResourceNone       ResourceType = "NONE"
	ResourceFlight     ResourceType = "FLIGHT"
	ResourceHotel      ResourceType = "HOTEL"
	ResourceGiftCard   ResourceType = "GIFT_CARD"
	ResourceMerchant   ResourceType = "MERCHANT"
	ResourceMovie      ResourceType = "MOVIE"
	ResourceBillIssuer ResourceType = "BILL_ISSUER"
)

// OfferType describes how price and commercial terms are resolved.
type OfferType string

const (
	OfferFixedPrice    OfferType = "FIXED_PRICE"
	OfferRatePlan      OfferType = "RATE_PLAN"
	OfferRealtimeQuote OfferType = "REALTIME_QUOTE"
	OfferBillQuery     OfferType = "BILL_QUERY"
)

// AvailabilityType describes where sellability is checked.
type AvailabilityType string

const (
	AvailabilityUnlimited        AvailabilityType = "UNLIMITED"
	AvailabilityLocalPool        AvailabilityType = "LOCAL_POOL"
	AvailabilityRealtimeSupplier AvailabilityType = "REALTIME_SUPPLIER"
	AvailabilitySeatMap          AvailabilityType = "SEAT_MAP"
)

// BookingMode describes whether an order needs resource locking.
type BookingMode string

const (
	BookingNone            BookingMode = "NONE"
	BookingPreLock         BookingMode = "PRE_LOCK"
	BookingPayThenLock     BookingMode = "PAY_THEN_LOCK"
	BookingConfirmAfterPay BookingMode = "CONFIRM_AFTER_PAY"
)

// FulfillmentType describes the final delivery action.
type FulfillmentType string

const (
	FulfillmentTopup          FulfillmentType = "TOPUP"
	FulfillmentBillPay        FulfillmentType = "BILL_PAY"
	FulfillmentIssueCode      FulfillmentType = "ISSUE_CODE"
	FulfillmentTicket         FulfillmentType = "TICKET"
	FulfillmentBookingConfirm FulfillmentType = "BOOKING_CONFIRM"
)

// SupplierDependencyLevel describes how much the category depends on suppliers at runtime.
type SupplierDependencyLevel string

const (
	SupplierDependencyLow    SupplierDependencyLevel = "LOW"
	SupplierDependencyMedium SupplierDependencyLevel = "MEDIUM"
	SupplierDependencyHigh   SupplierDependencyLevel = "HIGH"
)

// Scene describes where the runtime context is being built.
type Scene string

const (
	SceneList        Scene = "list"
	SceneDetail      Scene = "detail"
	SceneCheckout    Scene = "checkout"
	SceneCreateOrder Scene = "create_order"
)

// CategoryCapability is the category capability matrix used to select runtime behavior.
type CategoryCapability struct {
	CategoryID              int64
	CategoryName            string
	ProductModelType        ProductModelType
	ResourceType            ResourceType
	OfferType               OfferType
	AvailabilityType        AvailabilityType
	InputSchemaID           string
	BookingMode             BookingMode
	FulfillmentType         FulfillmentType
	RefundRuleID            string
	SupplierDependencyLevel SupplierDependencyLevel
}
