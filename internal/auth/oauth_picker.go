package auth

import (
	"bytes"
	"fmt"
	"html/template"
)

// PickerProviders are shown on the local / page (tray login without --provider).
// Only Google is offered in the HTML UI; other providers remain available via
// tray login --provider <id> when enabled in Supabase.
var PickerProviders = []struct {
	ID    string
	Label string
}{
	{"google", "Continue with Google"},
}

type pickerLink struct {
	Href  string
	Label string
}

func buildPickerLinks(projectURL, redirectTo, codeChallenge string) ([]pickerLink, error) {
	out := make([]pickerLink, 0, len(PickerProviders))
	for _, p := range PickerProviders {
		href, err := AuthorizeURL(projectURL, p.ID, redirectTo, codeChallenge)
		if err != nil {
			return nil, err
		}
		out = append(out, pickerLink{Href: href, Label: p.Label})
	}
	return out, nil
}

var pickerPageTmpl = template.Must(template.New("picker").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Tray — Sign in</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 22rem; margin: 2.5rem auto; padding: 0 1rem; color: #111; }
  h1 { font-size: 1.25rem; font-weight: 600; margin-bottom: 0.5rem; }
  p { color: #444; font-size: 0.9rem; margin-bottom: 1.25rem; line-height: 1.4; }
  a {
    display: block; margin: 0.5rem 0; padding: 0.75rem 1rem;
    background: #111; color: #fff; text-decoration: none; border-radius: 8px; text-align: center; font-size: 0.95rem;
  }
  a:hover { background: #333; }
</style>
</head>
<body>
<h1>Sign in to Tray</h1>
<p>Sign in with Google. Ensure Google is enabled under Supabase → Authentication → Providers.</p>
{{range .Links}}
<a href="{{.Href}}">{{.Label}}</a>
{{end}}
</body>
</html>
`))

func renderPickerPage(links []pickerLink) (string, error) {
	var buf bytes.Buffer
	if err := pickerPageTmpl.Execute(&buf, map[string]any{"Links": links}); err != nil {
		return "", fmt.Errorf("auth: render picker: %w", err)
	}
	return buf.String(), nil
}
