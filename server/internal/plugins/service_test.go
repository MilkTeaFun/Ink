package plugins

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ruhuang/ink/server/internal/auth"
)

type memoryRepo struct {
	mu            sync.Mutex
	installations map[string]Installation
	bindings      map[string]Binding
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		installations: map[string]Installation{},
		bindings:      map[string]Binding{},
	}
}

func (r *memoryRepo) ListInstallations(_ context.Context) ([]Installation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]Installation, 0, len(r.installations))
	for _, installation := range r.installations {
		result = append(result, installation)
	}
	return result, nil
}

func (r *memoryRepo) FindInstallationByID(_ context.Context, installationID string) (*Installation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	installation, exists := r.installations[installationID]
	if !exists {
		return nil, nil
	}
	copy := installation
	return &copy, nil
}

func (r *memoryRepo) FindInstallationByPluginKey(_ context.Context, pluginKey string) (*Installation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, installation := range r.installations {
		if installation.PluginKey == pluginKey {
			copy := installation
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *memoryRepo) SaveInstallation(_ context.Context, installation Installation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.installations[installation.ID] = installation
	return nil
}

func (r *memoryRepo) ListPluginBindingsByUserID(_ context.Context, userID string) ([]Binding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]Binding, 0, len(r.bindings))
	for _, binding := range r.bindings {
		if binding.UserID == userID {
			result = append(result, binding)
		}
	}
	return result, nil
}

func (r *memoryRepo) FindPluginBindingByInstallationAndUserID(_ context.Context, installationID string, userID string) (*Binding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, binding := range r.bindings {
		if binding.PluginInstallationID == installationID && binding.UserID == userID {
			copy := binding
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *memoryRepo) SavePluginBinding(_ context.Context, binding Binding) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.bindings[binding.ID] = binding
	return nil
}

type fakeAuthenticator struct{}

func (fakeAuthenticator) GetCurrentUser(_ context.Context, accessToken string) (auth.UserDTO, error) {
	switch accessToken {
	case "admin-token":
		return auth.UserDTO{
			ID:    "admin-user",
			Email: "admin",
			Name:  "Administrator",
			Role:  "admin",
		}, nil
	case "member-token":
		return auth.UserDTO{
			ID:    "member-user",
			Email: "member",
			Name:  "Member",
			Role:  "member",
		}, nil
	default:
		return auth.UserDTO{}, errors.New("invalid token")
	}
}

type fakeEncryptor struct{}

func (fakeEncryptor) Encrypt(plaintext string) ([]byte, []byte, error) {
	return []byte(plaintext), []byte("nonce"), nil
}

func (fakeEncryptor) Decrypt(ciphertext []byte, _ []byte) (string, error) {
	return string(ciphertext), nil
}

type fakeIDGenerator struct {
	mu      sync.Mutex
	counter int
}

func (g *fakeIDGenerator) New(prefix string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.counter++
	return prefix + "-" + time.Now().Format("150405") + "-" + string(rune('a'+g.counter)), nil
}

type fakeClock struct {
	now time.Time
}

func (c fakeClock) Now() time.Time {
	return c.now
}

func TestUploadPluginSaveBindingAndExecuteFetch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := newMemoryRepo()
	pluginRoot := t.TempDir()
	fixtureDir := filepath.Join("..", "..", "testdata", "plugins", "node-hello-plugin")
	zipPath := filepath.Join(t.TempDir(), "node-hello-plugin.zip")
	if err := zipDirectory(fixtureDir, zipPath, true); err != nil {
		t.Fatalf("zip fixture: %v", err)
	}

	service := NewService(
		repo,
		fakeAuthenticator{},
		fakeEncryptor{},
		&fakeIDGenerator{},
		fakeClock{now: time.Date(2026, 4, 10, 2, 0, 0, 0, time.UTC)},
		nil,
		pluginRoot,
		5*time.Second,
		30*time.Second,
	)

	file, err := os.Open(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	uploaded, err := service.UploadPlugin(ctx, "admin-token", "node-hello-plugin.zip", file)
	if err != nil {
		t.Fatalf("upload plugin: %v", err)
	}

	if uploaded.Installation.Status != InstallationStatusReady {
		t.Fatalf("expected ready installation, got %s", uploaded.Installation.Status)
	}
	if uploaded.Installation.PluginKey != "node-hello-source" {
		t.Fatalf("unexpected plugin key: %s", uploaded.Installation.PluginKey)
	}

	installation, manifest, err := service.GetInstallation(ctx, uploaded.Installation.ID)
	if err != nil {
		t.Fatalf("get installation: %v", err)
	}
	if manifest.Runtime.Type != "node" {
		t.Fatalf("expected node runtime, got %s", manifest.Runtime.Type)
	}
	if _, err := os.Stat(filepath.Join(installation.CurrentPath, "ink-plugin.json")); err != nil {
		t.Fatalf("expected installed manifest: %v", err)
	}

	saved, err := service.SaveBinding(ctx, "member-token", installation.ID, BindingInput{
		Enabled: true,
		Config: map[string]any{
			"sourceName":         "Fixture Source",
			"includeTriggeredAt": true,
		},
		Secrets: map[string]string{
			"apiToken": "super-secret",
		},
	})
	if err != nil {
		t.Fatalf("save binding: %v", err)
	}
	if saved.Binding == nil || saved.Binding.Status != BindingStatusConnected {
		t.Fatalf("expected connected binding, got %+v", saved.Binding)
	}

	validation, err := service.TestBinding(ctx, "member-token", installation.ID, BindingInput{
		Enabled: true,
		Config: map[string]any{
			"sourceName": "Fixture Source",
		},
	})
	if err != nil {
		t.Fatalf("test binding: %v", err)
	}
	if !validation.Valid {
		t.Fatalf("expected valid binding, got %+v", validation)
	}

	binding, secrets, err := service.GetBindingForUser(ctx, installation.ID, "member-user")
	if err != nil {
		t.Fatalf("get binding: %v", err)
	}
	if secrets["apiToken"] != "super-secret" {
		t.Fatalf("expected decrypted secret, got %+v", secrets)
	}

	result, err := service.ExecuteFetch(ctx, installation, binding, secrets, map[string]any{
		"message": "Hello plugin",
		"tone":    "verbose",
		"repeat":  2,
	}, FetchTrigger{
		ScheduledFor: "2026-04-10T10:00:00+08:00",
		TriggeredAt:  "2026-04-10T02:00:00Z",
		Timezone:     "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("execute fetch: %v", err)
	}
	if result.Title != "Fixture Source Digest" {
		t.Fatalf("unexpected title: %s", result.Title)
	}
	if !strings.Contains(result.Content, "Hello plugin\nHello plugin") {
		t.Fatalf("unexpected content: %s", result.Content)
	}
	if !strings.Contains(result.Content, "Triggered at: 2026-04-10T02:00:00Z") {
		t.Fatalf("expected triggered time in content: %s", result.Content)
	}
}

func TestUnzipSecureRejectsPathTraversal(t *testing.T) {
	t.Parallel()

	zipPath := filepath.Join(t.TempDir(), "invalid.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}

	writer := zip.NewWriter(file)
	entry, err := writer.Create("../escape.txt")
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := entry.Write([]byte("escape")); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}

	err = unzipSecure(zipPath, t.TempDir())
	if !errors.Is(err, ErrInvalidPlugin) {
		t.Fatalf("expected invalid plugin error, got %v", err)
	}
}

func TestUnzipSecureRejectsArchiveExceedingActualLimit(t *testing.T) {
	t.Parallel()

	zipPath := filepath.Join(t.TempDir(), "oversized.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}

	writer := zip.NewWriter(file)
	entry, err := writer.Create("large.txt")
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := entry.Write([]byte("0123456789")); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}

	err = unzipSecureWithLimit(zipPath, t.TempDir(), 8)
	if !errors.Is(err, ErrInvalidPlugin) {
		t.Fatalf("expected invalid plugin error, got %v", err)
	}
}

func zipDirectory(sourceDir string, zipPath string, wrapTopLevel bool) error {
	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	writer := zip.NewWriter(file)

	basePrefix := ""
	if wrapTopLevel {
		basePrefix = filepath.Base(sourceDir)
	}

	if err := filepath.WalkDir(sourceDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == sourceDir {
			return nil
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		zipName := filepath.ToSlash(relativePath)
		if basePrefix != "" {
			zipName = filepath.ToSlash(filepath.Join(basePrefix, relativePath))
		}

		if entry.IsDir() {
			_, err := writer.Create(zipName + "/")
			return err
		}

		header, err := zip.FileInfoHeader(mustStat(path))
		if err != nil {
			return err
		}
		header.Name = zipName
		header.Method = zip.Deflate

		target, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		source, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = source.Close()
		}()

		_, err = io.Copy(target, source)
		return err
	}); err != nil {
		_ = writer.Close()
		return err
	}

	return writer.Close()
}

func mustStat(path string) os.FileInfo {
	info, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	return info
}
