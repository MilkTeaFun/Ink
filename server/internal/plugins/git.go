package plugins

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// safeGitURLPattern matches the narrow set of HTTPS URLs the installer will
// pass to the git CLI: https://<host>/<path>. Hosts are alphanumerics plus
// "." and "-"; paths are alphanumerics plus "._-/". This is re-checked on
// the normalized URL inside validateGitURL so static analysis can see that
// the value reaching exec.CommandContext is sanitized.
var safeGitURLPattern = regexp.MustCompile(`^https://[A-Za-z0-9][A-Za-z0-9.\-]*(?::[0-9]+)?/[A-Za-z0-9._\-/]+$`)

// safeGitRefPattern restricts the branch/tag name to characters that cannot
// be interpreted as arguments or shell metacharacters. This is intentionally
// stricter than git's own rules.
var safeGitRefPattern = regexp.MustCompile(`^[A-Za-z0-9._\-/]+$`)

// GitCloner performs a shallow clone of a remote git repository into destDir
// and returns the resolved commit SHA. Implementations must never prompt for
// credentials interactively; network transports other than HTTPS should be
// rejected before Clone is called.
type GitCloner interface {
	Clone(ctx context.Context, repoURL, ref, destDir string) (commitSHA string, err error)
}

// ExecGitCloner shells out to the `git` binary. It is used by the real
// service wiring in production; tests substitute a fake.
type ExecGitCloner struct{}

// Clone runs `git clone --depth 1 --single-branch [--branch <ref>] <url> <dest>`
// with interactive credential prompts disabled, then resolves HEAD to a commit
// SHA. An empty ref lets the remote default branch be cloned.
//
// The caller MUST have already routed repoURL through validateGitURL and ref
// (when non-empty) must match safeGitRefPattern; as a defense in depth we
// re-check both here so that any future caller cannot accidentally pass an
// unsanitized value to exec.CommandContext.
func (ExecGitCloner) Clone(ctx context.Context, repoURL, ref, destDir string) (string, error) {
	if !safeGitURLPattern.MatchString(repoURL) {
		return "", fmt.Errorf("%w: repository URL failed safety check", ErrInvalidInput)
	}
	ref = strings.TrimSpace(ref)
	if ref != "" && !safeGitRefPattern.MatchString(ref) {
		return "", fmt.Errorf("%w: git ref contains unsupported characters", ErrInvalidInput)
	}

	args := []string{"clone", "--depth", "1", "--single-branch"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, "--", repoURL, destDir)

	cloneEnv := append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=/bin/echo",
		"GCM_INTERACTIVE=never",
	)

	cloneCmd := exec.CommandContext(ctx, "git", args...) //nolint:gosec // repoURL + ref sanitized via safeGitURLPattern / safeGitRefPattern above
	cloneCmd.Env = cloneEnv
	var cloneStderr bytes.Buffer
	cloneCmd.Stderr = &cloneStderr
	if err := cloneCmd.Run(); err != nil {
		return "", fmt.Errorf("git clone: %w: %s", err, strings.TrimSpace(cloneStderr.String()))
	}

	revCmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	revCmd.Dir = destDir
	var revStdout bytes.Buffer
	var revStderr bytes.Buffer
	revCmd.Stdout = &revStdout
	revCmd.Stderr = &revStderr
	if err := revCmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse: %w: %s", err, strings.TrimSpace(revStderr.String()))
	}

	return strings.TrimSpace(revStdout.String()), nil
}

// GitInstallInput describes a request to install a plugin from a remote git
// repository. Ref and Subdir are both optional.
type GitInstallInput struct {
	RepoURL string `json:"repoUrl"`
	Ref     string `json:"ref"`
	Subdir  string `json:"subdir"`
}

// DefaultGitAllowedHosts enumerates the hosts the service accepts when the
// operator has not customized PLUGIN_GIT_ALLOWED_HOSTS. The list is
// intentionally small — SSRF-style probing against internal services is the
// biggest risk on this path.
var DefaultGitAllowedHosts = []string{
	"github.com",
	"gitee.com",
	"gitlab.com",
}

// validateGitURL parses and sanity-checks the provided repository URL against
// the allowlist. It returns a normalized URL (scheme + host + path) that can
// be handed to git clone.
func validateGitURL(rawURL string, allowedHosts []string) (string, string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", "", fmt.Errorf("%w: repository URL is required", ErrInvalidInput)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("%w: invalid repository URL", ErrInvalidInput)
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return "", "", fmt.Errorf("%w: repository URL must use https", ErrInvalidInput)
	}
	if parsed.User != nil {
		return "", "", fmt.Errorf("%w: credentials must not be embedded in the URL", ErrInvalidInput)
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return "", "", fmt.Errorf("%w: repository URL is missing a host", ErrInvalidInput)
	}
	if !isHostAllowed(host, allowedHosts) {
		return "", "", fmt.Errorf("%w: repository host %q is not allowed", ErrInvalidInput, host)
	}

	// Reconstruct the URL without query/fragment so callers can't smuggle in
	// credentials or auth params via the request.
	normalized := &url.URL{
		Scheme: strings.ToLower(parsed.Scheme),
		Host:   parsed.Host,
		Path:   parsed.Path,
	}
	normalizedStr := normalized.String()
	if !safeGitURLPattern.MatchString(normalizedStr) {
		return "", "", fmt.Errorf("%w: repository URL contains unsupported characters", ErrInvalidInput)
	}
	return normalizedStr, host, nil
}

// isHostAllowed reports whether host matches any entry in allowed. Entries
// may be exact hostnames ("github.com") or a single-level wildcard
// ("*.example.com") to permit sub-domains.
func isHostAllowed(host string, allowed []string) bool {
	for _, entry := range allowed {
		entry = strings.TrimSpace(strings.ToLower(entry))
		if entry == "" {
			continue
		}
		if entry == host {
			return true
		}
		if strings.HasPrefix(entry, "*.") {
			suffix := entry[1:] // ".example.com"
			if strings.HasSuffix(host, suffix) && len(host) > len(suffix) {
				return true
			}
		}
	}
	return false
}

// sanitizeSubdir validates that subdir is a well-formed relative path that
// stays inside the clone directory.
func sanitizeSubdir(subdir string) (string, error) {
	subdir = strings.TrimSpace(subdir)
	if subdir == "" {
		return "", nil
	}
	if filepath.IsAbs(subdir) {
		return "", fmt.Errorf("%w: subdir must be a relative path", ErrInvalidInput)
	}
	cleaned := filepath.Clean(subdir)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: subdir must not escape the repository root", ErrInvalidInput)
	}
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("%w: subdir must be a relative path", ErrInvalidInput)
	}
	return cleaned, nil
}

// resolvePluginDirectoryInClone returns the directory inside cloneDir that
// contains ink-plugin.json. If subdir is set it is applied as a relative
// path; otherwise the clone root (or a single top-level subdirectory) is
// searched.
func resolvePluginDirectoryInClone(cloneDir string, subdir string) (string, error) {
	if subdir != "" {
		pluginDir := filepath.Join(cloneDir, subdir)
		rel, err := filepath.Rel(cloneDir, pluginDir)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return "", fmt.Errorf("%w: subdir escapes the repository", ErrInvalidInput)
		}
		info, err := os.Stat(pluginDir)
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("%w: subdir %q not found in repository", ErrInvalidPlugin, subdir)
		}
		if _, err := os.Stat(filepath.Join(pluginDir, "ink-plugin.json")); err != nil {
			return "", fmt.Errorf("%w: ink-plugin.json missing in subdir %q", ErrInvalidPlugin, subdir)
		}
		return pluginDir, nil
	}
	return resolvePluginDirectory(cloneDir)
}

// ErrGitInstallDisabled is returned by InstallFromGit when the service was
// constructed without a GitCloner.
var ErrGitInstallDisabled = errors.New("git install is not enabled on this server")
