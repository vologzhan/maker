package dto

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Uuid      uuid.UUID
	CreatedAt time.Time
	DeletedAt *time.Time

	// maker:keep-dto-relations
}
