package util

import (
	"time"

	"path/filepath"
	"regexp"

	"github.com/satori/go.uuid"
)

type (
	Resource struct {
		ID        uuid.UUID `db:"id"`
		Resource  string    `db:"resource"`
		FileName  string    `db:"file_name"`
		CreatedAt time.Time `db:"created_at"`
	}

	ResourceStorage interface {
		Persist(r Resource) error
		FindAllByResource(resource string) ([]Resource, error)
		FindLastByResourceAndFilesGroup(resource, wName string) (Resource, error)
	}
)

func NewResource(resource, fname string) Resource {
	return Resource{
		ID:        uuid.NewV4(),
		Resource:  resource,
		FileName:  fname,
		CreatedAt: time.Now(),
	}
}

func GeResourceNameFromFilePath(file string) string {
	var dir = filepath.Dir(file)
	var ext = filepath.Ext(file)

	name := file[len(dir)+1 : len(file)-len(ext)]

	reg := regexp.MustCompile(`(^[0-9]+_)+(.*)`)
	res := reg.ReplaceAllString(name, "${2}")

	return res
}
