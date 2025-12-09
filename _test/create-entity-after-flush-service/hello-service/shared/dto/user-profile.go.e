package dto

import "github.com/google/uuid"

type UserProfile struct {
	UserUuid   uuid.UUID
	Name       string
	SecondName *string

	// maker:keep-dto-relations
}
