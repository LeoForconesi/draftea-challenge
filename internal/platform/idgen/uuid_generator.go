package idgen

import "github.com/google/uuid"

// UUIDGenerator creates random UUIDs.
type UUIDGenerator struct{}

func (UUIDGenerator) New() uuid.UUID {
	return uuid.New()
}
