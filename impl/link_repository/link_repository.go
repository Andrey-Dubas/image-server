package link_repository

import (
	"time"

	"github.com/google/uuid"
)

type ILinkRepository interface {
	GenerateLink(ttl time.Duration) (uuid.UUID, error)
	HasLink(uuid.UUID) bool
}
