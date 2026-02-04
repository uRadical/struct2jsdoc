package examples

import (
	"time"
)

// Order represents a customer order with various data types
type Order struct {
	ID            string           `json:"id"`
	OrderNumber   string           `json:"orderNumber"`             // Human-readable order number
	CustomerID    string           `json:"customerId"`              // Customer ID
	Status        string           `json:"status"`                  // Order status (pending, processing, shipped, delivered)
	Items         []OrderItem      `json:"items"`                   // Order items
	Subtotal      float64          `json:"subtotal"`                // Subtotal before tax and shipping
	Tax           float64          `json:"tax"`                     // Tax amount
	Shipping      float64          `json:"shipping"`                // Shipping cost
	Total         float64          `json:"total"`                   // Total amount
	Currency      string           `json:"currency"`                // Currency code
	ShippingAddr  Address          `json:"shippingAddress"`         // Shipping address
	BillingAddr   *Address         `json:"billingAddress,omitempty"` // Billing address (optional, defaults to shipping)
	PaymentMethod string           `json:"paymentMethod"`           // Payment method
	CreatedAt     time.Time        `json:"createdAt"`               // Order creation time
	UpdatedAt     time.Time        `json:"updatedAt"`               // Last update time
	ShippedAt     *time.Time       `json:"shippedAt,omitempty"`     // Shipping time
	DeliveredAt   *time.Time       `json:"deliveredAt,omitempty"`   // Delivery time
	Notes         string           `json:"notes,omitempty"`         // Order notes
	Discounts     []Discount       `json:"discounts,omitempty"`     // Applied discounts
}

// OrderItem represents a single item in an order
type OrderItem struct {
	ProductID string  `json:"productId"`           // Product ID
	Name      string  `json:"name"`                // Product name
	Quantity  int     `json:"quantity"`            // Quantity ordered
	Price     float64 `json:"price"`               // Price per unit
	Total     float64 `json:"total"`               // Total price (quantity * price)
	SKU       string  `json:"sku,omitempty"`       // Stock keeping unit
}

// Address represents a physical address
type Address struct {
	Line1      string `json:"line1"`               // Address line 1
	Line2      string `json:"line2,omitempty"`     // Address line 2
	City       string `json:"city"`                // City
	State      string `json:"state,omitempty"`     // State/Province
	PostalCode string `json:"postalCode"`          // Postal/ZIP code
	Country    string `json:"country"`             // Country code (e.g., US, CA, GB)
}

// Discount represents a discount applied to an order
type Discount struct {
	Code   string  `json:"code"`              // Discount code
	Type   string  `json:"type"`              // Discount type (percentage, fixed)
	Amount float64 `json:"amount"`            // Discount amount
	Reason string  `json:"reason,omitempty"`  // Reason for discount
}

// Statistics contains various numeric statistics
type Statistics struct {
	TotalOrders     uint64             `json:"totalOrders"`     // Total number of orders
	TotalRevenue    float64            `json:"totalRevenue"`    // Total revenue
	AverageOrder    float64            `json:"averageOrder"`    // Average order value
	TopProducts     []string           `json:"topProducts"`     // Top selling product IDs
	RevenueByMonth  map[string]float64 `json:"revenueByMonth"`  // Revenue grouped by month
	OrdersByStatus  map[string]int     `json:"ordersByStatus"`  // Order counts by status
}
