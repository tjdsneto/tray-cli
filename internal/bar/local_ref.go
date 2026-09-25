package bar

import (
	"strings"

	"github.com/tjdsneto/tray-cli/internal/localtray"
)

// LocalTrayRef returns the first clipboard token for a local tray.
// Global trays use the bare name; dir/branch use the canonical tray id (unambiguous for agents).
func LocalTrayRef(rec localtray.TrayRecord) string {
	switch rec.Kind {
	case localtray.KindGlobal:
		return strings.TrimSpace(rec.Name)
	default:
		return strings.TrimSpace(rec.ID)
	}
}
