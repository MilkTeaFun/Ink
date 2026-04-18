package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ruhuang/ink/server/internal/platform/secret"
)

func TestEnsureDevEnvGeneratesValidSecrets(t *testing.T) {
	t.Parallel()

	requireOpenSSL(t)
	values := generateDevEnv(t)

	assertGeneratedValue(t, values, "JWT_SECRET", "replace-with-a-long-random-secret")
	aiKey := assertGeneratedValue(t, values, "AI_CONFIG_ENCRYPTION_KEY", "replace-with-a-32-byte-base64-or-raw-secret")
	assertValidSecretKey(t, aiKey)
}

func parseDotEnv(raw string) map[string]string {
	values := make(map[string]string)

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		values[key] = value
	}

	return values
}

func requireOpenSSL(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("openssl is required for ensure_dev_env.sh")
	}
}

func generateDevEnv(t *testing.T) map[string]string {
	t.Helper()

	template := mustReadFile(t, filepath.Join("..", ".env.example"))
	script := mustReadFile(t, "ensure_dev_env.sh")

	tempDir := t.TempDir()
	serverDir := filepath.Join(tempDir, "server")
	scriptsDir := filepath.Join(serverDir, "scripts")

	mustMkdirAll(t, scriptsDir, 0o755)
	mustWriteFile(t, filepath.Join(serverDir, ".env.example"), template, 0o644)
	mustWriteFile(t, filepath.Join(scriptsDir, "ensure_dev_env.sh"), script, 0o755)
	mustRunScript(t, filepath.Join(scriptsDir, "ensure_dev_env.sh"))

	return parseDotEnv(string(mustReadFile(t, filepath.Join(serverDir, ".env"))))
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return content
}

func mustMkdirAll(t *testing.T, path string, perm os.FileMode) {
	t.Helper()

	if err := os.MkdirAll(path, perm); err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, content []byte, perm os.FileMode) {
	t.Helper()

	if err := os.WriteFile(path, content, perm); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustRunScript(t *testing.T, path string) {
	t.Helper()

	cmd := exec.Command("sh", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run %s: %v\n%s", path, err, output)
	}
}

func assertGeneratedValue(t *testing.T, values map[string]string, key string, placeholder string) string {
	t.Helper()

	value := values[key]
	if value == "" || value == placeholder {
		t.Fatalf("expected generated %s, got %q", key, value)
	}

	return value
}

func assertValidSecretKey(t *testing.T, value string) {
	t.Helper()

	if _, err := secret.NewBox(value); err != nil {
		t.Fatalf("generated AI_CONFIG_ENCRYPTION_KEY is invalid: %v", err)
	}
}
