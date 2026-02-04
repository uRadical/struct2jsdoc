package examples

import "time"

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`                  // Unique user identifier
	Username  string    `json:"username"`            // User's login name
	Email     string    `json:"email"`               // User's email address
	FirstName string    `json:"firstName,omitempty"` // User's first name
	LastName  string    `json:"lastName,omitempty"`  // User's last name
	Age       int       `json:"age,omitempty"`       // User's age
	IsActive  bool      `json:"isActive"`            // Whether the user account is active
	CreatedAt time.Time `json:"createdAt"`           // Account creation timestamp
	UpdatedAt time.Time `json:"updatedAt"`           // Last update timestamp
	Profile   *Profile  `json:"profile,omitempty"`   // User's profile information
	Tags      []string  `json:"tags,omitempty"`      // User tags
}

// Profile contains additional user profile information
type Profile struct {
	Bio       string            `json:"bio,omitempty"`       // User biography
	AvatarURL string            `json:"avatarUrl,omitempty"` // URL to user's avatar image
	Website   string            `json:"website,omitempty"`   // User's website
	Location  string            `json:"location,omitempty"`  // User's location
	Metadata  map[string]string `json:"metadata,omitempty"`  // Additional metadata
}
