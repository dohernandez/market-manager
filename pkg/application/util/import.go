package util

import (
	"time"

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
		FindLastByResourceAndWallet(resource, wName string) (Resource, error)
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
