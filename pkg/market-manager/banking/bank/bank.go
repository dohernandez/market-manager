package bank

import (
	"github.com/satori/go.uuid"
)

type AccountNoType string

type Account struct {
	ID        uuid.UUID
	Name      string
	AccountNo string
	Alias     string
}

const (
	IBAN AccountNoType = "iban"
)
