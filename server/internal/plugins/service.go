package plugins

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ruhuang/ink/server/internal/auth"
)

var (
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidInput    = errors.New("invalid plugin input")
	ErrInvalidPlugin   = errors.New("invalid plugin package")
	ErrNotFound        = errors.New("plugin not found")
	ErrExecutionFailed = errors.New("plugin execution failed")
	ErrMissingSecret   = errors.New("plugin encryption secret missing")
)

type ValidationFailure struct {
	Errors []FieldError
}

func (v ValidationFailure) Error() string {
	if len(v.Errors) == 0 {
		return "插件配置校验失败"
	}

	return fmt.Sprintf("%s: %s", v.Errors[0].Field, v.Errors[0].Message)
}

type Repository interface {
	ListInstallations(ctx context.Context) ([]Installation, error)
	FindInstallationByID(ctx context.Context, installationID string) (*Installation, error)
	FindInstallationByPluginKey(ctx context.Context, pluginKey string) (*Installation, error)
	SaveInstallation(ctx context.Context, installation Installation) error
	ListPluginBindingsByUserID(ctx context.Context, userID string) ([]Binding, error)
	FindPluginBindingByInstallationAndUserID(ctx context.Context, installationID string, userID string) (*Binding, error)
	SavePluginBinding(ctx context.Context, binding Binding) error
}

type Authenticator interface {
	GetCurrentUser(ctx context.Context, accessToken string) (auth.UserDTO, error)
}

type Encryptor interface {
	Encrypt(plaintext string) ([]byte, []byte, error)
	Decrypt(ciphertext []byte, nonce []byte) (string, error)
}

type IDGenerator interface {
	New(prefix string) (string, error)
}

type Clock interface {
	Now() time.Time
}

type Runner interface {
	Run(ctx context.Context, workdir string, command []string, stdin []byte) ([]byte, []byte, error)
}

type Service struct {
	repo           Repository
	auth           Authenticator
	encryptor      Encryptor
	ids            IDGenerator
	clock          Clock
	runner         Runner
	pluginRoot     string
	execTimeout    time.Duration
	installTimeout time.Duration
}

type validationPayload struct {
	WorkspaceConfig map[string]any    `json:"workspaceConfig"`
	Secrets         map[string]string `json:"secrets"`
}

type fetchPayload struct {
	WorkspaceConfig map[string]any    `json:"workspaceConfig"`
	Secrets         map[string]string `json:"secrets"`
	ScheduleConfig  map[string]any    `json:"scheduleConfig"`
	Trigger         FetchTrigger      `json:"trigger"`
}

type execRunner struct{}

func NewService(
	repo Repository,
	authenticator Authenticator,
	encryptor Encryptor,
	ids IDGenerator,
	clock Clock,
	runner Runner,
	pluginRoot string,
	execTimeout time.Duration,
	installTimeout time.Duration,
) *Service {
	if runner == nil {
		runner = execRunner{}
	}

	return &Service{
		repo:           repo,
		auth:           authenticator,
		encryptor:      encryptor,
		ids:            ids,
		clock:          clock,
		runner:         runner,
		pluginRoot:     pluginRoot,
		execTimeout:    execTimeout,
		installTimeout: installTimeout,
	}
}

func (execRunner) Run(ctx context.Context, workdir string, command []string, stdin []byte) ([]byte, []byte, error) {
	if len(command) == 0 {
		return nil, nil, fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = workdir
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func (s *Service) ListAdminInstallations(ctx context.Context, accessToken string) ([]PluginDetails, error) {
	if err := s.requireAdmin(ctx, accessToken); err != nil {
		return nil, err
	}

	installations, err := s.repo.ListInstallations(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(installations, func(i, j int) bool {
		return installations[i].UpdatedAt.After(installations[j].UpdatedAt)
	})

	result := make([]PluginDetails, 0, len(installations))
	for _, installation := range installations {
		result = append(result, s.detailsFromInstallation(installation, nil))
	}

	return result, nil
}

func (s *Service) UploadPlugin(ctx context.Context, accessToken string, filename string, source io.Reader) (PluginDetails, error) {
	currentUser, err := s.requireAdminUser(ctx, accessToken)
	if err != nil {
		return PluginDetails{}, err
	}
	if err := s.ensurePluginRoot(); err != nil {
		return PluginDetails{}, err
	}

	stagingDir, err := os.MkdirTemp(s.pluginRoot, "plugin-upload-*")
	if err != nil {
		return PluginDetails{}, err
	}
	defer func() {
		_ = os.RemoveAll(stagingDir)
	}()

	zipPath := filepath.Join(stagingDir, sanitizedUploadName(filename))
	file, err := os.Create(zipPath)
	if err != nil {
		return PluginDetails{}, err
	}
	if _, err := io.Copy(file, source); err != nil {
		_ = file.Close()
		return PluginDetails{}, err
	}
	if err := file.Close(); err != nil {
		return PluginDetails{}, err
	}

	extractedDir := filepath.Join(stagingDir, "extracted")
	if err := os.MkdirAll(extractedDir, 0o755); err != nil {
		return PluginDetails{}, err
	}

	if err := unzipSecure(zipPath, extractedDir); err != nil {
		return PluginDetails{}, err
	}

	pluginDir, err := resolvePluginDirectory(extractedDir)
	if err != nil {
		return PluginDetails{}, err
	}

	manifestJSON, err := os.ReadFile(filepath.Join(pluginDir, "ink-plugin.json"))
	if err != nil {
		return PluginDetails{}, fmt.Errorf("%w: missing ink-plugin.json", ErrInvalidPlugin)
	}

	manifest, err := ParseManifest(manifestJSON)
	if err != nil {
		return PluginDetails{}, err
	}

	if err := validateRuntimeFiles(pluginDir, manifest); err != nil {
		return PluginDetails{}, err
	}

	if err := s.installPlugin(ctx, pluginDir, manifest); err != nil {
		existing, lookupErr := s.repo.FindInstallationByPluginKey(ctx, manifest.PluginKey)
		if lookupErr == nil && existing == nil {
			now := s.clock.Now()
			installationID, idErr := s.ids.New("plugin")
			if idErr == nil {
				installErr := err.Error()
				_ = s.repo.SaveInstallation(ctx, Installation{
					ID:           installationID,
					PluginKey:    manifest.PluginKey,
					SourceType:   SourceTypeUpload,
					DisplayName:  manifest.Name,
					Version:      manifest.Version,
					RuntimeType:  manifest.Runtime.Type,
					ManifestJSON: manifestJSON,
					Status:       InstallationStatusFailed,
					LastError:    &installErr,
					InstalledBy:  &currentUser.ID,
					CreatedAt:    now,
					UpdatedAt:    now,
				})
			}
		}
		return PluginDetails{}, err
	}

	existing, err := s.repo.FindInstallationByPluginKey(ctx, manifest.PluginKey)
	if err != nil {
		return PluginDetails{}, err
	}

	now := s.clock.Now()
	installationID := ""
	createdAt := now
	if existing != nil {
		installationID = existing.ID
		createdAt = existing.CreatedAt
	} else {
		installationID, err = s.ids.New("plugin")
		if err != nil {
			return PluginDetails{}, err
		}
	}

	finalDir := filepath.Join(s.pluginRoot, "installations", fmt.Sprintf("%s-%d", installationID, now.UnixNano()))
	if err := os.MkdirAll(filepath.Dir(finalDir), 0o755); err != nil {
		return PluginDetails{}, err
	}
	if err := os.Rename(pluginDir, finalDir); err != nil {
		return PluginDetails{}, err
	}

	installation := Installation{
		ID:           installationID,
		PluginKey:    manifest.PluginKey,
		SourceType:   SourceTypeUpload,
		DisplayName:  manifest.Name,
		Version:      manifest.Version,
		RuntimeType:  manifest.Runtime.Type,
		ManifestJSON: manifestJSON,
		CurrentPath:  finalDir,
		Status:       InstallationStatusReady,
		InstalledBy:  &currentUser.ID,
		CreatedAt:    createdAt,
		UpdatedAt:    now,
	}
	if err := s.repo.SaveInstallation(ctx, installation); err != nil {
		return PluginDetails{}, err
	}

	return s.detailsFromInstallation(installation, nil), nil
}

func (s *Service) DisableInstallation(ctx context.Context, accessToken string, installationID string) (PluginDetails, error) {
	if err := s.requireAdmin(ctx, accessToken); err != nil {
		return PluginDetails{}, err
	}

	installation, manifest, err := s.GetInstallation(ctx, installationID)
	if err != nil {
		return PluginDetails{}, err
	}

	installation.Status = InstallationStatusDisabled
	installation.UpdatedAt = s.clock.Now()
	installation.LastError = nil
	if err := s.repo.SaveInstallation(ctx, installation); err != nil {
		return PluginDetails{}, err
	}

	details := s.detailsFromInstallation(installation, nil)
	details.Manifest = manifest
	return details, nil
}

func (s *Service) ListUserPlugins(ctx context.Context, accessToken string) ([]PluginDetails, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	installations, err := s.repo.ListInstallations(ctx)
	if err != nil {
		return nil, err
	}
	bindings, err := s.repo.ListPluginBindingsByUserID(ctx, currentUser.ID)
	if err != nil {
		return nil, err
	}

	bindingByInstallation := map[string]Binding{}
	for _, binding := range bindings {
		bindingByInstallation[binding.PluginInstallationID] = binding
	}

	result := make([]PluginDetails, 0, len(installations))
	for _, installation := range installations {
		if installation.Status == InstallationStatusFailed || installation.Status == InstallationStatusInstalling {
			continue
		}
		binding, hasBinding := bindingByInstallation[installation.ID]
		if hasBinding {
			result = append(result, s.detailsFromInstallation(installation, &binding))
			continue
		}
		result = append(result, s.detailsFromInstallation(installation, nil))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Installation.DisplayName < result[j].Installation.DisplayName
	})

	return result, nil
}

func (s *Service) GetUserPlugin(ctx context.Context, accessToken string, installationID string) (PluginDetails, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return PluginDetails{}, err
	}

	installation, manifest, err := s.GetInstallation(ctx, installationID)
	if err != nil {
		return PluginDetails{}, err
	}

	binding, err := s.repo.FindPluginBindingByInstallationAndUserID(ctx, installation.ID, currentUser.ID)
	if err != nil {
		return PluginDetails{}, err
	}

	details := s.detailsFromInstallation(installation, binding)
	details.Manifest = manifest
	return details, nil
}

func (s *Service) SaveBinding(ctx context.Context, accessToken string, installationID string, input BindingInput) (PluginDetails, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return PluginDetails{}, err
	}

	installation, manifest, err := s.GetInstallation(ctx, installationID)
	if err != nil {
		return PluginDetails{}, err
	}
	if installation.Status != InstallationStatusReady && input.Enabled {
		return PluginDetails{}, fmt.Errorf("%w: plugin is not ready", ErrInvalidInput)
	}

	existing, err := s.repo.FindPluginBindingByInstallationAndUserID(ctx, installation.ID, currentUser.ID)
	if err != nil {
		return PluginDetails{}, err
	}

	baseConfig := map[string]any{}
	existingSecrets := map[string]string{}
	if existing != nil {
		baseConfig = cloneMap(existing.Config)
		existingSecrets, err = s.decryptSecrets(*existing)
		if err != nil {
			return PluginDetails{}, err
		}
	}
	for key, value := range input.Config {
		baseConfig[key] = value
	}

	normalizedConfig, incomingSecrets, fieldErrs := NormalizeConfigValues(manifest.WorkspaceConfigSchema, baseConfig, true)
	if len(fieldErrs) > 0 {
		return PluginDetails{}, ValidationFailure{Errors: fieldErrs}
	}

	mergedSecrets := mergeSecrets(existingSecrets, input.Secrets, incomingSecrets)
	if input.Enabled {
		validation, err := s.runValidation(ctx, installation, normalizedConfig, mergedSecrets, manifest)
		if err != nil {
			return PluginDetails{}, err
		}
		if !validation.Valid {
			return PluginDetails{}, ValidationFailure{Errors: validation.Errors}
		}
	}

	now := s.clock.Now()
	binding := Binding{
		PluginInstallationID: installation.ID,
		UserID:               currentUser.ID,
		Enabled:              input.Enabled,
		Config:               normalizedConfig,
		Status:               BindingStatusDisconnected,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if existing != nil {
		binding.ID = existing.ID
		binding.CreatedAt = existing.CreatedAt
		binding.Status = existing.Status
	}
	if binding.ID == "" {
		binding.ID, err = s.ids.New("binding")
		if err != nil {
			return PluginDetails{}, err
		}
	}

	if len(mergedSecrets) > 0 {
		if s.encryptor == nil {
			return PluginDetails{}, ErrMissingSecret
		}
		ciphertext, nonce, err := s.encryptSecrets(mergedSecrets)
		if err != nil {
			return PluginDetails{}, err
		}
		binding.Ciphertext = ciphertext
		binding.Nonce = nonce
	}

	if input.Enabled {
		binding.Status = BindingStatusConnected
		binding.LastValidatedAt = &now
		binding.LastError = nil
	} else {
		binding.Status = BindingStatusDisconnected
		binding.LastValidatedAt = nil
		binding.LastError = nil
	}

	if err := s.repo.SavePluginBinding(ctx, binding); err != nil {
		return PluginDetails{}, err
	}

	details := s.detailsFromInstallation(installation, &binding)
	details.Manifest = manifest
	return details, nil
}

func (s *Service) TestBinding(ctx context.Context, accessToken string, installationID string, input BindingInput) (ValidationResult, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return ValidationResult{}, err
	}

	installation, manifest, err := s.GetInstallation(ctx, installationID)
	if err != nil {
		return ValidationResult{}, err
	}
	if installation.Status != InstallationStatusReady {
		return ValidationResult{}, fmt.Errorf("%w: plugin is not ready", ErrInvalidInput)
	}

	existing, err := s.repo.FindPluginBindingByInstallationAndUserID(ctx, installation.ID, currentUser.ID)
	if err != nil {
		return ValidationResult{}, err
	}

	baseConfig := map[string]any{}
	existingSecrets := map[string]string{}
	if existing != nil {
		baseConfig = cloneMap(existing.Config)
		existingSecrets, err = s.decryptSecrets(*existing)
		if err != nil {
			return ValidationResult{}, err
		}
	}
	for key, value := range input.Config {
		baseConfig[key] = value
	}

	normalizedConfig, incomingSecrets, fieldErrs := NormalizeConfigValues(manifest.WorkspaceConfigSchema, baseConfig, true)
	if len(fieldErrs) > 0 {
		return ValidationResult{
			Valid:  false,
			Errors: fieldErrs,
		}, nil
	}

	mergedSecrets := mergeSecrets(existingSecrets, input.Secrets, incomingSecrets)
	return s.runValidation(ctx, installation, normalizedConfig, mergedSecrets, manifest)
}

func (s *Service) GetInstallation(ctx context.Context, installationID string) (Installation, Manifest, error) {
	installation, err := s.repo.FindInstallationByID(ctx, installationID)
	if err != nil {
		return Installation{}, Manifest{}, err
	}
	if installation == nil {
		return Installation{}, Manifest{}, ErrNotFound
	}

	manifest, err := ParseManifest(installation.ManifestJSON)
	if err != nil {
		return Installation{}, Manifest{}, err
	}

	return *installation, manifest, nil
}

func (s *Service) GetBindingForUser(ctx context.Context, installationID string, userID string) (Binding, map[string]string, error) {
	binding, err := s.repo.FindPluginBindingByInstallationAndUserID(ctx, installationID, userID)
	if err != nil {
		return Binding{}, nil, err
	}
	if binding == nil {
		return Binding{}, nil, ErrNotFound
	}

	secrets, err := s.decryptSecrets(*binding)
	if err != nil {
		return Binding{}, nil, err
	}

	return *binding, secrets, nil
}

func (s *Service) ExecuteFetch(ctx context.Context, installation Installation, binding Binding, secrets map[string]string, scheduleConfig map[string]any, trigger FetchTrigger) (FetchResult, error) {
	manifest, err := ParseManifest(installation.ManifestJSON)
	if err != nil {
		return FetchResult{}, err
	}

	payload := fetchPayload{
		WorkspaceConfig: cloneMap(binding.Config),
		Secrets:         secrets,
		ScheduleConfig:  cloneMap(scheduleConfig),
		Trigger:         trigger,
	}
	input, err := json.Marshal(payload)
	if err != nil {
		return FetchResult{}, err
	}

	execCtx, cancel := context.WithTimeout(ctx, s.execTimeout)
	defer cancel()

	stdout, stderr, err := s.runner.Run(execCtx, installation.CurrentPath, manifest.Entrypoints.Fetch.Command, input)
	if err != nil {
		return FetchResult{}, fmt.Errorf("%w: %s", ErrExecutionFailed, trimExecOutput(stdout, stderr, err))
	}

	var result FetchResult
	if err := json.Unmarshal(stdout, &result); err != nil {
		return FetchResult{}, fmt.Errorf("%w: invalid fetch output", ErrExecutionFailed)
	}
	if strings.TrimSpace(result.Title) == "" || strings.TrimSpace(result.Content) == "" {
		return FetchResult{}, fmt.Errorf("%w: fetch output must include title and content", ErrExecutionFailed)
	}
	if strings.TrimSpace(result.SourceLabel) == "" {
		result.SourceLabel = installation.DisplayName
	}

	return result, nil
}

func (s *Service) runValidation(ctx context.Context, installation Installation, config map[string]any, secrets map[string]string, manifest Manifest) (ValidationResult, error) {
	payload := validationPayload{
		WorkspaceConfig: cloneMap(config),
		Secrets:         secrets,
	}
	input, err := json.Marshal(payload)
	if err != nil {
		return ValidationResult{}, err
	}

	execCtx, cancel := context.WithTimeout(ctx, s.execTimeout)
	defer cancel()

	stdout, stderr, err := s.runner.Run(execCtx, installation.CurrentPath, manifest.Entrypoints.Validate.Command, input)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("%w: %s", ErrExecutionFailed, trimExecOutput(stdout, stderr, err))
	}

	var result ValidationResult
	if err := json.Unmarshal(stdout, &result); err != nil {
		return ValidationResult{}, fmt.Errorf("%w: invalid validation output", ErrExecutionFailed)
	}

	if !result.Valid && len(result.Errors) == 0 {
		result.Errors = []FieldError{{
			Field:   "",
			Message: "插件校验失败",
		}}
	}

	return result, nil
}

func (s *Service) installPlugin(ctx context.Context, pluginDir string, manifest Manifest) error {
	var command []string
	switch manifest.Runtime.Type {
	case "node":
		command = []string{"pnpm", "install", "--frozen-lockfile"}
	case "python":
		command = []string{"uv", "sync", "--frozen"}
	default:
		return fmt.Errorf("%w: unsupported runtime %s", ErrInvalidPlugin, manifest.Runtime.Type)
	}

	installCtx, cancel := context.WithTimeout(ctx, s.installTimeout)
	defer cancel()

	stdout, stderr, err := s.runner.Run(installCtx, pluginDir, command, nil)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrExecutionFailed, trimExecOutput(stdout, stderr, err))
	}

	return nil
}

func (s *Service) detailsFromInstallation(installation Installation, binding *Binding) PluginDetails {
	manifest, _ := ParseManifest(installation.ManifestJSON)
	details := PluginDetails{
		Installation: InstallationSummary{
			ID:          installation.ID,
			PluginKey:   installation.PluginKey,
			SourceType:  installation.SourceType,
			DisplayName: installation.DisplayName,
			Version:     installation.Version,
			RuntimeType: installation.RuntimeType,
			Status:      installation.Status,
			Description: manifest.Description,
			CreatedAt:   &installation.CreatedAt,
			UpdatedAt:   &installation.UpdatedAt,
		},
		Manifest: manifest,
	}
	if installation.LastError != nil {
		details.Installation.LastError = *installation.LastError
	}

	if binding != nil {
		details.Binding = &BindingSummary{
			ID:              binding.ID,
			Enabled:         binding.Enabled,
			Status:          binding.Status,
			Config:          cloneMap(binding.Config),
			LastValidatedAt: binding.LastValidatedAt,
		}
		if binding.LastError != nil {
			details.Binding.LastError = *binding.LastError
		}
	}

	return details
}

func (s *Service) requireAdmin(ctx context.Context, accessToken string) error {
	_, err := s.requireAdminUser(ctx, accessToken)
	return err
}

func (s *Service) requireAdminUser(ctx context.Context, accessToken string) (auth.UserDTO, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return auth.UserDTO{}, err
	}
	if currentUser.Role != "admin" {
		return auth.UserDTO{}, ErrForbidden
	}
	return currentUser, nil
}

func (s *Service) ensurePluginRoot() error {
	return os.MkdirAll(s.pluginRoot, 0o755)
}

func (s *Service) encryptSecrets(secrets map[string]string) ([]byte, []byte, error) {
	if len(secrets) == 0 {
		return nil, nil, nil
	}
	if s.encryptor == nil {
		return nil, nil, ErrMissingSecret
	}

	payload, err := json.Marshal(secrets)
	if err != nil {
		return nil, nil, err
	}
	return s.encryptor.Encrypt(string(payload))
}

func (s *Service) decryptSecrets(binding Binding) (map[string]string, error) {
	if len(binding.Ciphertext) == 0 {
		return map[string]string{}, nil
	}
	if s.encryptor == nil {
		return nil, ErrMissingSecret
	}

	plaintext, err := s.encryptor.Decrypt(binding.Ciphertext, binding.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decrypt binding secrets: %w", err)
	}

	var secrets map[string]string
	if err := json.Unmarshal([]byte(plaintext), &secrets); err != nil {
		return nil, err
	}
	if secrets == nil {
		secrets = map[string]string{}
	}
	return secrets, nil
}

func mergeSecrets(existing map[string]string, inputs ...map[string]string) map[string]string {
	result := map[string]string{}
	for key, value := range existing {
		result[key] = value
	}
	for _, current := range inputs {
		for key, value := range current {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			result[key] = trimmed
		}
	}
	return result
}

func cloneMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func sanitizedUploadName(filename string) string {
	base := filepath.Base(strings.TrimSpace(filename))
	if base == "." || base == "" {
		return "plugin.zip"
	}
	if !strings.HasSuffix(strings.ToLower(base), ".zip") {
		return base + ".zip"
	}
	return base
}

func unzipSecure(zipPath string, destination string) error {
	const maxUncompressedBytes int64 = 64 << 20
	return unzipSecureWithLimit(zipPath, destination, maxUncompressedBytes)
}

func unzipSecureWithLimit(zipPath string, destination string, maxUncompressedBytes int64) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = reader.Close()
	}()

	var totalUncompressed int64
	destination = filepath.Clean(destination)

	for _, file := range reader.File {
		cleanName := filepath.Clean(file.Name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			return fmt.Errorf("%w: invalid zip entry path", ErrInvalidPlugin)
		}
		if file.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: symbolic links are not allowed", ErrInvalidPlugin)
		}

		targetPath := filepath.Join(destination, cleanName)
		relativePath, err := filepath.Rel(destination, targetPath)
		if err != nil || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("%w: invalid zip entry path", ErrInvalidPlugin)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		src, err := file.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			_ = src.Close()
			return err
		}

		remaining := maxUncompressedBytes - totalUncompressed
		if remaining < 0 {
			remaining = 0
		}

		written, err := io.Copy(dst, &io.LimitedReader{R: src, N: remaining + 1})
		totalUncompressed += written
		if err != nil {
			_ = dst.Close()
			_ = src.Close()
			return err
		}
		if totalUncompressed > maxUncompressedBytes {
			_ = dst.Close()
			_ = src.Close()
			return fmt.Errorf("%w: plugin archive is too large", ErrInvalidPlugin)
		}
		if err := dst.Close(); err != nil {
			_ = src.Close()
			return err
		}
		if err := src.Close(); err != nil {
			return err
		}
	}

	return nil
}

func resolvePluginDirectory(extractedDir string) (string, error) {
	rootManifest := filepath.Join(extractedDir, "ink-plugin.json")
	if _, err := os.Stat(rootManifest); err == nil {
		return extractedDir, nil
	}

	entries, err := os.ReadDir(extractedDir)
	if err != nil {
		return "", err
	}

	directories := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Name() == "__MACOSX" {
			continue
		}
		directories = append(directories, filepath.Join(extractedDir, entry.Name()))
	}
	if len(directories) == 1 {
		manifestPath := filepath.Join(directories[0], "ink-plugin.json")
		if _, err := os.Stat(manifestPath); err == nil {
			return directories[0], nil
		}
	}

	return "", fmt.Errorf("%w: plugin archive must contain ink-plugin.json at root or a single top-level directory", ErrInvalidPlugin)
}

func validateRuntimeFiles(pluginDir string, manifest Manifest) error {
	switch manifest.Runtime.Type {
	case "node":
		if !fileExists(filepath.Join(pluginDir, "package.json")) || !fileExists(filepath.Join(pluginDir, "pnpm-lock.yaml")) {
			return fmt.Errorf("%w: node plugins must include package.json and pnpm-lock.yaml", ErrInvalidPlugin)
		}
	case "python":
		if !fileExists(filepath.Join(pluginDir, "pyproject.toml")) || !fileExists(filepath.Join(pluginDir, "uv.lock")) {
			return fmt.Errorf("%w: python plugins must include pyproject.toml and uv.lock", ErrInvalidPlugin)
		}
	default:
		return fmt.Errorf("%w: unsupported runtime %s", ErrInvalidPlugin, manifest.Runtime.Type)
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func trimExecOutput(stdout []byte, stderr []byte, runErr error) string {
	parts := []string{}
	if text := strings.TrimSpace(string(stdout)); text != "" {
		parts = append(parts, text)
	}
	if text := strings.TrimSpace(string(stderr)); text != "" {
		parts = append(parts, text)
	}
	if runErr != nil && len(parts) == 0 {
		parts = append(parts, runErr.Error())
	}
	return strings.Join(parts, " | ")
}
