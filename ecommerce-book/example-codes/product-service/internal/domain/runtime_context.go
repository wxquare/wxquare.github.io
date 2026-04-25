package domain

// Money stores monetary values in cents.
type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// ProductDefinition is the stable product identity used before transaction.
type ProductDefinition struct {
	SKUID      int64             `json:"sku_id"`
	SPUID      int64             `json:"spu_id,omitempty"`
	SKUCode    string            `json:"sku_code"`
	Title      string            `json:"title"`
	CategoryID int64             `json:"category_id"`
	BasePrice  Money             `json:"base_price"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// ResourceContext describes the underlying resource, such as hotel, route, or card brand.
type ResourceContext struct {
	ResourceType ResourceType      `json:"resource_type"`
	ResourceID   string            `json:"resource_id,omitempty"`
	Name         string            `json:"name,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

// OfferContext describes price and commercial rules for the current scene.
type OfferContext struct {
	OfferType   OfferType         `json:"offer_type"`
	OfferID     string            `json:"offer_id,omitempty"`
	RatePlanID  string            `json:"rate_plan_id,omitempty"`
	Price       Money             `json:"price"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	NeedRefresh bool              `json:"need_refresh"`
}

// AvailabilityContext describes whether the product can be sold.
type AvailabilityContext struct {
	AvailabilityType AvailabilityType  `json:"availability_type"`
	Sellable         bool              `json:"sellable"`
	Quantity         int               `json:"quantity,omitempty"`
	Message          string            `json:"message,omitempty"`
	Attributes       map[string]string `json:"attributes,omitempty"`
}

// InputSchema describes user inputs needed before checkout or order creation.
type InputSchema struct {
	SchemaID string       `json:"schema_id"`
	Fields   []InputField `json:"fields"`
}

type InputField struct {
	Name        string `json:"name"`
	Label       string `json:"label,omitempty"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Pattern     string `json:"pattern,omitempty"`
	Description string `json:"description,omitempty"`
}

// BookingRequirement describes whether the platform must lock external resources.
type BookingRequirement struct {
	Mode         BookingMode       `json:"mode"`
	Required     bool              `json:"required"`
	TTLSeconds   int               `json:"ttl_seconds,omitempty"`
	ProviderSide bool              `json:"provider_side"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

// FulfillmentContract describes how a paid order will be delivered.
type FulfillmentContract struct {
	Type       FulfillmentType   `json:"type"`
	Mode       string            `json:"mode"`
	TimeoutSec int               `json:"timeout_sec,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// RefundRule describes after-sale boundaries.
type RefundRule struct {
	RuleID      string            `json:"rule_id"`
	Refundable  bool              `json:"refundable"`
	NeedReview  bool              `json:"need_review"`
	Description string            `json:"description"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// ProductRuntimeContext is the eight-layer product transaction context.
type ProductRuntimeContext struct {
	SKUID             int64                `json:"sku_id"`
	CategoryID        int64                `json:"category_id"`
	CategoryName      string               `json:"category_name"`
	Scene             Scene                `json:"scene"`
	ProductDefinition *ProductDefinition   `json:"product_definition"`
	Resource          *ResourceContext     `json:"resource"`
	Offer             *OfferContext        `json:"offer"`
	Availability      *AvailabilityContext `json:"availability"`
	InputSchema       *InputSchema         `json:"input_schema"`
	Booking           *BookingRequirement  `json:"booking"`
	Fulfillment       *FulfillmentContract `json:"fulfillment"`
	RefundRule        *RefundRule          `json:"refund_rule"`
}

func (c *ProductRuntimeContext) IsComplete() bool {
	return c != nil &&
		c.ProductDefinition != nil &&
		c.Resource != nil &&
		c.Offer != nil &&
		c.Availability != nil &&
		c.InputSchema != nil &&
		c.Booking != nil &&
		c.Fulfillment != nil &&
		c.RefundRule != nil
}
