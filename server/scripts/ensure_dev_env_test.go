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

	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("openssl is required for ensure_dev_env.sh")
	}

	template, err := os.ReadFile(filepath.Join("..", ".env.example"))
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	script, err := os.ReadFile("ensure_dev_env.sh")
	if err != nil {
		t.Fatalf("read script: %v", err)
	}

	tempDir := t.TempDir()
	serverDir := filepath.Join(tempDir, "server")
	scriptsDir := filepath.Join(serverDir, "scripts")

	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("create scripts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(serverDir, ".env.example"), template, 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "ensure_dev_env.sh"), script, 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	cmd := exec.Command("sh", filepath.Join(scriptsDir, "ensure_dev_env.sh"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run ensure_dev_env.sh: %v\n%s", err, output)
	}

	envFile := filepath.Join(serverDir, ".env")
	envContent, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("read generated env: %v", err)
	}

	values := parseDotEnv(string(envContent))
	if got := values["JWT_SECRET"]; got == "" || got == "replace-with-a-long-random-secret" {
		t.Fatalf("expected generated JWT_SECRET, got %q", got)
	}

	aiKey := values["AI_CONFIG_ENCRYPTION_KEY"]
	if aiKey == "" || aiKey == "replace-with-a-32-byte-base64-or-raw-secret" {
		t.Fatalf("expected generated AI_CONFIG_ENCRYPTION_KEY, got %q", aiKey)
	}

	if _, err := secret.NewBox(aiKey); err != nil {
		t.Fatalf("generated AI_CONFIG_ENCRYPTION_KEY is invalid: %v", err)
	}
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
