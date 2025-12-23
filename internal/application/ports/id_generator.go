package ports

import "github.com/google/uuid"

// IDGenerator abstracts UUID generation for testability.
type IDGenerator interface {
	New() uuid.UUID
}
