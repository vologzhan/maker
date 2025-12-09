package dto

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Uuid      uuid.UUID
	DeletedAt *time.Time

	// maker:keep-dto-relations
}
