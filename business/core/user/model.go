package user

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        mail.Address
	Roles        []Role
	PasswordHash []byte
	Enabled      bool
	PhoneNo      PhoneNumber
	DateCreated  time.Time
	DateUpdated  time.Time
}

type NewUser struct {
	Name            string
	Email           mail.Address
	Roles           []Role
	PhoneNo         PhoneNumber
	Password        string
	PasswordConfirm string
}

type UpdateUser struct {
	Name            *string
	Email           *mail.Address
	Roles           []Role
	Password        *string
	PasswordConfirm *string
	Enabled         *bool
}
