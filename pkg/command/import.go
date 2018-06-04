package command

import (
	"context"

	"path/filepath"

	"regexp"

	"path"

	"github.com/dohernandez/market-manager/pkg/container"
	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type (
	// ImportCommand ...
	ImportCommand struct {
		*BaseCommand
	}

	resourceImport struct {
		filePath     string
		resourceName string
	}
)

// NewImportCommand constructs ImportCommand
func NewImportCommand(baseCommand *BaseCommand) *ImportCommand {
	return &ImportCommand{
		BaseCommand: baseCommand,
	}
}

func (cmd *ImportCommand) runImport(
	ctx context.Context,
	c *container.Container,
	resourceType string,
	ris []resourceImport,
	fn func(ctx context.Context, c *container.Container, ri resourceImport) error,
) error {
	is := c.ImportStorageInstance()
	irs, err := is.FindAllByResource(resourceType)
	if err != nil {
		if err != mm.ErrNotFound {
			return err
		}

		irs = []_import.Resource{}
	}

	for _, ri := range ris {
		fileName := path.Base(ri.filePath)

		var found bool
		for _, ir := range irs {
			if ir.FileName == fileName {
				found = true

				break
			}
		}

		if !found {
			logger.FromContext(ctx).Infof("Importing file %s", fileName)

			if err := fn(ctx, c, ri); err != nil {
				return err
			}

			ir := _import.NewResource(resourceType, fileName)
			err := is.Persist(ir)
			if err != nil {
				return err
			}

			logger.FromContext(ctx).Infof("Imported file %s", fileName)
		}
	}

	return nil
}

func (cmd *ImportCommand) geResourceNameFromFilePath(file string) string {
	var dir = filepath.Dir(file)
	var ext = filepath.Ext(file)

	name := file[len(dir)+1 : len(file)-len(ext)]

	reg := regexp.MustCompile(`(^[0-9]+_)+(.*)`)
	res := reg.ReplaceAllString(name, "${2}")

	return res
}
