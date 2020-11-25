package events

import (
	"time"
)

type DeletedContent struct {
	Deleted struct {
		AtTime time.Time `json:"at_time"`
		// Stored but not in the view
		ByIdentityID string `json:"by_identity_id,omitempty"`
		// Set during view generation
		ByIdentity *SenderView `json:"by_identity,omitempty"`
	} `json:"deleted"`
}
