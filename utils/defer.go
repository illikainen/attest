package utils

import (
	"errors"
	"os"
)

// revive:disable-next-line
func DeferRemove(name string, err *error) { //nolint
	*err = errors.Join(*err, os.RemoveAll(name))
}
