package cmd_cli

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type ResourceImport struct {
	FilePath     string
	ResourceName string
}

type busExecuteContextFunc func(ctx context.Context, ri ResourceImport) (interface{}, error)

func Import(
	ctxt context.Context,
	busExecuteContext busExecuteContextFunc,
	resourceStorage util.ResourceStorage,
	resource,
	resourcePath,
	filePath,
	fileName string,
) error {
	log := logger.FromContext(ctxt)

	ris, err := getResourceImports(ctxt, filePath, resourcePath, fileName)
	if err != nil {
		log.WithError(err).Fatal("Failed importing")
	}

	fn := func(ctx context.Context, busExecuteContext busExecuteContextFunc, ri ResourceImport) error {
		_, err := busExecuteContext(ctx, ri)
		if err != nil {
			logger.FromContext(ctx).WithError(err).Fatal("Failed importing %s", ri.FilePath)
		}

		return nil
	}

	err = runImport(ctxt, busExecuteContext, resourceStorage, resource, ris, fn)
	if err != nil {
		log.WithError(err).Fatal("Failed importing")
	}

	log.Info("Import finished")

	return nil
}

func getResourceImports(_ context.Context, filePath, importPath, fileName string) ([]ResourceImport, error) {
	var ris []ResourceImport

	if filePath != "" {
		filePath := fmt.Sprintf("%s/%s.csv", importPath, filePath)

		ris = append(ris, ResourceImport{
			FilePath:     filePath,
			ResourceName: util.GeResourceNameFromFilePath(filePath),
		})
	} else {
		err := filepath.Walk(importPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".csv" {
				filePath := path
				rName := util.GeResourceNameFromFilePath(filePath)

				found := true

				if fileName != "" && fileName != rName {
					found = false
				}

				if found {
					ris = append(ris, ResourceImport{
						FilePath:     filePath,
						ResourceName: rName,
					})
				}
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return ris, nil
}

func runImport(
	ctxt context.Context,
	busExecuteContext busExecuteContextFunc,
	resourceStorage util.ResourceStorage,
	resourceType string,
	ris []ResourceImport,
	fn func(ctx context.Context, busExecuteContext busExecuteContextFunc, ri ResourceImport) error,
) error {
	irs, err := resourceStorage.FindAllByResource(resourceType)
	if err != nil {
		if err != mm.ErrNotFound {
			return err
		}

		irs = []util.Resource{}
	}

	log := logger.FromContext(ctxt)

	for _, ri := range ris {
		fileName := path.Base(ri.FilePath)

		var found bool
		for _, ir := range irs {
			if ir.FileName == fileName {
				found = true

				break
			}
		}

		if !found {
			log.Infof("Importing file %s", fileName)

			if err := fn(ctxt, busExecuteContext, ri); err != nil {
				return err
			}

			ir := util.NewResource(resourceType, fileName)
			err := resourceStorage.Persist(ir)
			if err != nil {
				return err
			}

			log.Infof("Imported file %s", fileName)
		}
	}

	return nil
}
