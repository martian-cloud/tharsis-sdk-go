package types

import (
	"time"
)

// ResourceMetadata contains metadata for a particular resource
//
// Keeping the ID field here rather than following
// GraphQL in pulling the ID field out to the parent.
type ResourceMetadata struct {
	CreationTimestamp    *time.Time `json:"createdAt"`
	LastUpdatedTimestamp *time.Time `json:"updatedAt,omitempty" `
	ID                   string     `json:"id"`
	Version              string     `json:"version"`
	TRN                  string     `json:"trn"`
}
