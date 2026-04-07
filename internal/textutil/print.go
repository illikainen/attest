package textutil

import (
	"fmt"
)

func Printf(format string, v ...any) {
	// revive:disable-next-line
	if _, err := fmt.Printf(format, v...); err != nil { //nolint
		panic(err)
	}
}
