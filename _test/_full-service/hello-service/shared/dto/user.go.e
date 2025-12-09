package dto

import "time"

type User struct {
	Id        int
	CreatedAt time.Time
	DeletedAt *time.Time

	// maker:keep-dto-relations
}
