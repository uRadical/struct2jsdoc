package examples

import "time"

// APIResponse is a generic API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`           // Whether the request was successful
	Data    interface{} `json:"data,omitempty"`    // Response data
	Error   *APIError   `json:"error,omitempty"`   // Error information if request failed
	Meta    *MetaData   `json:"meta,omitempty"`    // Additional metadata
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`              // Error code
	Message string `json:"message"`           // Human-readable error message
	Details string `json:"details,omitempty"` // Additional error details
}

// MetaData contains pagination and other metadata
type MetaData struct {
	Page       int `json:"page"`                // Current page number
	PerPage    int `json:"perPage"`             // Items per page
	Total      int `json:"total"`               // Total number of items
	TotalPages int `json:"totalPages"`          // Total number of pages
}

// struct2jsdoc: nogen
// InternalConfig should not be exported to JSDoc
type InternalConfig struct {
	Secret    string `json:"-"`
	APIKey    string `json:"-"`
	DebugMode bool   `json:"debugMode"`
}

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`                // Event type
	Timestamp time.Time              `json:"timestamp"`           // When the event occurred
	UserID    string                 `json:"userId,omitempty"`    // User who triggered the event
	Data      map[string]interface{} `json:"data"`                // Event data
	Tags      []string               `json:"tags,omitempty"`      // Event tags
}

// PaginatedResponse represents a paginated list response
type PaginatedResponse struct {
	Items      []interface{} `json:"items"`      // List of items
	Pagination MetaData      `json:"pagination"` // Pagination information
}
