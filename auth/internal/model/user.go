package model

import "time"

type NotificationMethod struct {
	ProviderName string
	Target       string
}

type User struct {
	UUID                string
	Login               string
	Email               string
	PasswordHash        string
	NotificationMethods []NotificationMethod
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}

// RegisterInfo is the data needed to register a user.
type RegisterInfo struct {
	Login               string
	Password            string
	Email               string
	NotificationMethods []NotificationMethod
}
