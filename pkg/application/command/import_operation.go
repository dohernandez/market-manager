package command

import "github.com/satori/go.uuid"

type ImportOperation struct {
	FilePath string
	Wallet   string
	Trades   map[uuid.UUID]string
}
