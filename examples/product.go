package examples

// Product represents a product in an e-commerce system
type Product struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`                // Product name
	Description string   `json:"description"`         // Product description
	Price       float64  `json:"price"`               // Product price in USD
	Currency    string   `json:"currency"`            // Currency code (e.g., USD, EUR)
	InStock     bool     `json:"inStock"`             // Whether product is in stock
	SKU         string   `json:"sku,omitempty"`       // Stock keeping unit
	Categories  []string `json:"categories"`          // Product categories
	Images      []Image  `json:"images,omitempty"`    // Product images
	Variants    []Variant `json:"variants,omitempty"` // Product variants
	Dimensions  *Dimensions `json:"dimensions,omitempty"` // Product dimensions
}

// Image represents a product image
type Image struct {
	URL       string `json:"url"`                 // Image URL
	Alt       string `json:"alt,omitempty"`       // Alternative text
	Width     int    `json:"width,omitempty"`     // Image width in pixels
	Height    int    `json:"height,omitempty"`    // Image height in pixels
	IsPrimary bool   `json:"isPrimary,omitempty"` // Whether this is the primary image
}

// Variant represents a product variant (e.g., different sizes or colors)
type Variant struct {
	ID    string            `json:"id"`
	Name  string            `json:"name"`               // Variant name
	SKU   string            `json:"sku"`                // Variant SKU
	Price float64           `json:"price,omitempty"`    // Variant-specific price
	Attributes map[string]string `json:"attributes"` // Variant attributes (e.g., color, size)
}

// Dimensions represents product dimensions
type Dimensions struct {
	Length float64 `json:"length"` // Length in inches
	Width  float64 `json:"width"`  // Width in inches
	Height float64 `json:"height"` // Height in inches
	Weight float64 `json:"weight"` // Weight in pounds
}
