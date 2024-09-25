package user

import (
	"errors"
	"fmt"

	"github.com/nyaruka/phonenumbers"
)

type PhoneNumber struct {
	number string
}

func ParsePhoneNumber(value string) (PhoneNumber, error) {
	pn, err := phonenumbers.Parse(value, "US")
	if err != nil {
		return PhoneNumber{}, fmt.Errorf("parsing phone number: %w", err)
	}

	if !phonenumbers.IsValidNumber(pn) {
		return PhoneNumber{}, errors.New("invalid phone number")
	}

	formatted := phonenumbers.Format(pn, phonenumbers.E164)
	return PhoneNumber{number: formatted}, nil
}

func (n PhoneNumber) Number() string {
	return n.number
}

func (n *PhoneNumber) UnmarshalText(data []byte) error {
	num, err := ParsePhoneNumber(string(data))
	if err != nil {
		return err
	}

	n.number = num.number
	return nil
}

func (n PhoneNumber) MarshalText() ([]byte, error) {
	return []byte(n.number), nil
}

func (n PhoneNumber) Equal(n2 PhoneNumber) bool {
	return n.number == n2.number
}
