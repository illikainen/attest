package git

import (
	"fmt"
	"strings"
)

type Submodule struct {
	Name        string
	Initialized bool
}

func SubmoduleStatus(dir string) ([]*Submodule, error) {
	p, err := Git(dir, "submodule", "status", "--recursive")
	if err != nil {
		return nil, err
	}

	var submodules []*Submodule
	for _, line := range strings.Split(string(p.Stdout), "\n") {
		if line == "" {
			continue
		}

		elts := strings.Split(strings.TrimLeft(line, " "), " ")
		if len(elts) != 2 && len(elts) != 3 {
			return nil, fmt.Errorf("invalid submodule status line: %s", line)
		}

		submodules = append(submodules, &Submodule{
			Name:        elts[1],
			Initialized: !strings.HasPrefix(line, "-"),
		})
	}

	return submodules, nil
}
