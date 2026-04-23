package plugins

import (
	"net/url"
	"testing"
)

func FuzzValidateGitURL(f *testing.F) {
	for _, rawURL := range []string{
		"https://github.com/owner/repo",
		"https://gitlab.com/group/repo.git?ref=main#fragment",
		"http://github.com/owner/repo",
		"https://token@github.com/owner/repo",
		"https://example.com/repo",
		"not a url",
		"",
	} {
		f.Add(rawURL)
	}

	allowedHosts := append([]string(nil), DefaultGitAllowedHosts...)

	f.Fuzz(func(t *testing.T, rawURL string) {
		normalized, host, err := validateGitURL(rawURL, allowedHosts)
		if err != nil {
			return
		}

		parsed, parseErr := url.Parse(normalized)
		if parseErr != nil {
			t.Fatalf("normalized URL must stay parseable: %v", parseErr)
		}
		if parsed.Scheme != "https" {
			t.Fatalf("normalized URL must stay on https, got %q", parsed.Scheme)
		}
		if parsed.User != nil {
			t.Fatalf("normalized URL must not contain user info: %q", normalized)
		}
		if parsed.RawQuery != "" || parsed.Fragment != "" {
			t.Fatalf("normalized URL must drop query and fragment: %q", normalized)
		}
		if parsed.Hostname() != host {
			t.Fatalf("normalized host mismatch: %q != %q", parsed.Hostname(), host)
		}
		if !isHostAllowed(host, allowedHosts) {
			t.Fatalf("normalized host must stay in allowlist: %q", host)
		}
	})
}
