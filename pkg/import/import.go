package _import

import (
	"time"

	"github.com/satori/go.uuid"
)

type (
	Import interface {
		Import() error
	}

	Resource struct {
		ID        uuid.UUID
		Resource  string
		FileName  string
		CreatedAt time.Time
	}

	Storage interface {
		Persist(r Resource) error
		FindAllByResource(resource string) ([]Resource, error)
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
