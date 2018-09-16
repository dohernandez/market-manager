package mm

import "github.com/pkg/errors"

// ErrNotFound is a common not found error
var ErrNotFound = errors.New("Not found")

// ErrInvalidIDFormat means that the ID format is not valid
var ErrInvalidIDFormat = errors.New("Invalid ID format. Should be of type uuid, e.g.:\"c0c48a8f-fa6b-4467-9fd3-3bfc96ed44bb\"")

// ErrCanNotAddOperation means that the operation can not be added to the wallet
var ErrCanNotAddOperation = errors.Errorf("can not add operation wallet item not found")
