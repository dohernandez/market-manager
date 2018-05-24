package bank

import (
	"github.com/almerlucke/go-iban/iban"
	"github.com/satori/go.uuid"
)

type Account struct {
	ID    uuid.UUID
	Name  string
	IBAN  *iban.IBAN
	Alias string
}
