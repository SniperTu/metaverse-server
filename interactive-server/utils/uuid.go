package utils

import "github.com/google/uuid"

func NewUUIDV4() string {
	return uuid.New().String()
}
