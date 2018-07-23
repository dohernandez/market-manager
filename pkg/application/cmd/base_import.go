package cmd

import (
	"context"
	"path"
	"path/filepath"
	"regexp"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type (
	// baseImport ...
	baseImport struct {
		resourceStorage util.ResourceStorage
	}

	resourceImport struct {
		filePath     string
		resourceName string
	}
)

func (cmd *baseImport) runImport(
	ctx context.Context,
	bus *cbus.Bus,
	resourceType string,
	ris []resourceImport,
	fn func(ctx context.Context, bus *cbus.Bus, ri resourceImport) error,
) error {
	irs, err := cmd.resourceStorage.FindAllByResource(resourceType)
	if err != nil {
		if err != mm.ErrNotFound {
			return err
		}

		irs = []util.Resource{}
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

			if err := fn(ctx, bus, ri); err != nil {
				return err
			}

			ir := util.NewResource(resourceType, fileName)
			err := cmd.resourceStorage.Persist(ir)
			if err != nil {
				return err
			}

			logger.FromContext(ctx).Infof("Imported file %s", fileName)
		}
	}

	return nil
}

func (cmd *baseImport) geResourceNameFromFilePath(file string) string {
	var dir = filepath.Dir(file)
	var ext = filepath.Ext(file)

	name := file[len(dir)+1 : len(file)-len(ext)]

	reg := regexp.MustCompile(`(^[0-9]+_)+(.*)`)
	res := reg.ReplaceAllString(name, "${2}")

	return res
}
