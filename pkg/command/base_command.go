package command

import (
	"context"

	"github.com/dohernandez/market-manager/pkg/config"
)

type (
	// BaseCommand hold common command properties
	BaseCommand struct {
		ctx    context.Context
		config *config.Specification
	}
)

// NewBaseCommand creates a structure with common shared properties of the commands
func NewBaseCommand(ctx context.Context, config *config.Specification) *BaseCommand {
	return &BaseCommand{
		ctx:    ctx,
		config: config,
	}
}
