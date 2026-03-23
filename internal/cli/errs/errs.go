// Package errs holds stable error values for CLI wiring without import cycles.
package errs

import "errors"

// MissingBackendConfig means URL/key (or equivalent) are not available to reach the server.
var MissingBackendConfig = errors.New("missing backend configuration")
