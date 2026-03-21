package cli

import "fmt"

func stub(name string) error {
	return fmt.Errorf("`tray %s` isn't available yet — we're still building it. Run `tray help` to see what you can use today", name)
}
