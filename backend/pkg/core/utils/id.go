package utils

import "github.com/google/uuid"

// NewID returns a new UUID v4 string for primary keys.
func NewID() string {
	return uuid.NewString()
}
