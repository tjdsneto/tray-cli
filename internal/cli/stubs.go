package cli

import "fmt"

func stub(name string) error {
	return fmt.Errorf("%s: not implemented yet", name)
}
