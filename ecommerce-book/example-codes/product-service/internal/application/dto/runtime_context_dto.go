package dto

type RuntimeContextResponse struct {
	SKUID             int64                  `json:"sku_id"`
	CategoryID        int64                  `json:"category_id"`
	CategoryName      string                 `json:"category_name"`
	Scene             string                 `json:"scene"`
	ProductDefinition ProductDefinitionDTO   `json:"product_definition"`
	Resource          ResourceContextDTO     `json:"resource"`
	Offer             OfferContextDTO        `json:"offer"`
	Availability      AvailabilityContextDTO `json:"availability"`
	InputSchema       InputSchemaDTO         `json:"input_schema"`
	Booking           BookingRequirementDTO  `json:"booking"`
	Fulfillment       FulfillmentContractDTO `json:"fulfillment"`
	RefundRule        RefundRuleDTO          `json:"refund_rule"`
}

type ProductDefinitionDTO struct {
	SKUID      int64             `json:"sku_id"`
	SPUID      int64             `json:"spu_id,omitempty"`
	SKUCode    string            `json:"sku_code"`
	Title      string            `json:"title"`
	CategoryID int64             `json:"category_id"`
	BasePrice  MoneyDTO          `json:"base_price"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type MoneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type ResourceContextDTO struct {
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id,omitempty"`
	Name         string            `json:"name,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

type OfferContextDTO struct {
	OfferType   string            `json:"offer_type"`
	OfferID     string            `json:"offer_id,omitempty"`
	RatePlanID  string            `json:"rate_plan_id,omitempty"`
	Price       MoneyDTO          `json:"price"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	NeedRefresh bool              `json:"need_refresh"`
}

type AvailabilityContextDTO struct {
	AvailabilityType string            `json:"availability_type"`
	Sellable         bool              `json:"sellable"`
	Quantity         int               `json:"quantity,omitempty"`
	Message          string            `json:"message,omitempty"`
	Attributes       map[string]string `json:"attributes,omitempty"`
}

type InputSchemaDTO struct {
	SchemaID string          `json:"schema_id"`
	Fields   []InputFieldDTO `json:"fields"`
}

type InputFieldDTO struct {
	Name        string `json:"name"`
	Label       string `json:"label,omitempty"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Pattern     string `json:"pattern,omitempty"`
	Description string `json:"description,omitempty"`
}

type BookingRequirementDTO struct {
	Mode         string            `json:"mode"`
	Required     bool              `json:"required"`
	TTLSeconds   int               `json:"ttl_seconds,omitempty"`
	ProviderSide bool              `json:"provider_side"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

type FulfillmentContractDTO struct {
	Type       string            `json:"type"`
	Mode       string            `json:"mode"`
	TimeoutSec int               `json:"timeout_sec,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type RefundRuleDTO struct {
	RuleID      string            `json:"rule_id"`
	Refundable  bool              `json:"refundable"`
	NeedReview  bool              `json:"need_review"`
	Description string            `json:"description"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}
