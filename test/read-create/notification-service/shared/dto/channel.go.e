package dto

import (
	"time"

	"github.com/google/uuid"
)

type Channel struct {
	Uuid         uuid.UUID
	RelationUuid uuid.UUID
	DeletedAt    *time.Time
}
