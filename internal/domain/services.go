package domain

// Services bundles domain entry points for the CLI (one place to pass dependencies).
type Services struct {
	Trays    TrayService
	Items    ItemService
	Profiles ProfileService
}
