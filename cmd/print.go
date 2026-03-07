package cmd

import (
	"fmt"
)

func printf(format string, v ...any) {
	// revive:disable-next-line
	if _, err := fmt.Printf(format, v...); err != nil { //nolint
		panic(err)
	}
}
