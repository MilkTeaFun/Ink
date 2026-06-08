package plugins

import (
	"errors"
	"fmt"
	"testing"
)

func TestParseManifestAcceptsValidV2Manifest(t *testing.T) {
	t.Parallel()

	raw := []byte(v2ManifestJSON(`"fetchPolicy": { "type": "fixed_interval", "minutes": 15 },`))

	manifest, err := ParseManifest(raw)
	if err != nil {
		t.Fatalf("expected valid manifest, got %v", err)
	}
	if manifest.SchemaVersion != 2 {
		t.Fatalf("unexpected schema version: %d", manifest.SchemaVersion)
	}
	if manifest.FetchPolicy.Type != FetchPolicyTypeFixedInterval {
		t.Fatalf("unexpected fetch policy: %+v", manifest.FetchPolicy)
	}
	if manifest.FetchPolicy.Minutes != 15 {
		t.Fatalf("unexpected fetch minutes: %d", manifest.FetchPolicy.Minutes)
	}
}

func TestParseManifestRejectsInvalidFetchPolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		fetchBlock string
	}{
		{name: "missing", fetchBlock: ""},
		{name: "unsupported type", fetchBlock: `"fetchPolicy": { "type": "cron", "minutes": 15 },`},
		{name: "non-positive interval", fetchBlock: `"fetchPolicy": { "type": "fixed_interval", "minutes": 0 },`},
	}

	for _, testCase := range testCases {
		_, err := ParseManifest([]byte(v2ManifestJSON(testCase.fetchBlock)))
		if !errors.Is(err, ErrInvalidPlugin) {
			t.Fatalf("expected invalid plugin error, got %v", err)
		}
	}
}

func TestParseManifestAcceptsPermissions(t *testing.T) {
	t.Parallel()

	raw := []byte(v2ManifestJSON(`"fetchPolicy": { "type": "fixed_interval", "minutes": 15 },
		"permissions": {
			"network": {
				"mode": "declared_hosts",
				"hosts": ["api.example.com", "*.example.org"]
			},
			"filesystem": {
				"temp": true
			},
			"installScripts": false
		},`))

	manifest, err := ParseManifest(raw)
	if err != nil {
		t.Fatalf("expected valid manifest, got %v", err)
	}
	if manifest.Permissions == nil || manifest.Permissions.Network == nil {
		t.Fatalf("expected parsed permissions, got %+v", manifest.Permissions)
	}
	if manifest.Permissions.Network.Mode != NetworkPermissionDeclaredHosts {
		t.Fatalf("unexpected network mode: %+v", manifest.Permissions.Network)
	}
}

func TestParseManifestRejectsInvalidPermissions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		permissions string
	}{
		{
			name: "unsupported network mode",
			permissions: `"permissions": {
				"network": { "mode": "private_lan" }
			},`,
		},
		{
			name: "missing declared hosts",
			permissions: `"permissions": {
				"network": { "mode": "declared_hosts" }
			},`,
		},
		{
			name: "host is url",
			permissions: `"permissions": {
				"network": { "mode": "declared_hosts", "hosts": ["https://api.example.com"] }
			},`,
		},
		{
			name: "duplicate host",
			permissions: `"permissions": {
				"network": { "mode": "declared_hosts", "hosts": ["api.example.com", "API.example.com"] }
			},`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseManifest([]byte(v2ManifestJSON(`"fetchPolicy": { "type": "fixed_interval", "minutes": 15 },
				` + testCase.permissions)))
			if !errors.Is(err, ErrInvalidPlugin) {
				t.Fatalf("expected invalid plugin error, got %v", err)
			}
		})
	}
}

func v2ManifestJSON(fetchBlock string) string {
	return fmt.Sprintf(`{
		"schemaVersion": 2,
		"kind": "source",
		"pluginKey": "demo-source",
		"name": "Demo Source",
		"version": "1.0.0",
		"description": "demo",
		"runtime": { "type": "node" },
		%s
		"entrypoints": {
			"validate": { "command": ["node", "validate.mjs"] },
			"fetch": { "command": ["node", "fetch.mjs"] }
		},
		"workspaceConfigSchema": []
	}`, fetchBlock)
}

func TestNormalizeConfigValuesSeparatesSecretsAndReportsUnknownFields(t *testing.T) {
	t.Parallel()

	normalized, secrets, errs := NormalizeConfigValues([]FieldSpec{
		{
			Key:      "title",
			Label:    "Title",
			Type:     FieldTypeText,
			Required: true,
		},
		{
			Key:      "token",
			Label:    "Token",
			Type:     FieldTypeSecret,
			Required: false,
		},
		{
			Key:          "repeat",
			Label:        "Repeat",
			Type:         FieldTypeNumber,
			Required:     false,
			DefaultValue: 1,
		},
		{
			Key:      "enabled",
			Label:    "Enabled",
			Type:     FieldTypeCheckbox,
			Required: false,
		},
	}, map[string]any{
		"title":   "Fixture Title",
		"token":   "super-secret",
		"repeat":  "2",
		"enabled": true,
		"extra":   "ignored",
	}, true)

	if len(errs) != 1 || errs[0].Field != "extra" {
		t.Fatalf("expected extra field error, got %+v", errs)
	}
	if normalized["title"] != "Fixture Title" {
		t.Fatalf("unexpected normalized title: %+v", normalized)
	}
	if normalized["repeat"] != 2 {
		t.Fatalf("unexpected normalized repeat: %+v", normalized["repeat"])
	}
	if normalized["enabled"] != true {
		t.Fatalf("unexpected normalized enabled: %+v", normalized["enabled"])
	}
	if normalized["token"] != nil {
		t.Fatalf("secret field must not remain in normalized config: %+v", normalized)
	}
	if secrets["token"] != "super-secret" {
		t.Fatalf("unexpected secrets map: %+v", secrets)
	}
}

func TestParseManifestRejectsSurroundingWhitespaceInIdentifiers(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"schemaVersion": 2,
		"kind": "source",
		"pluginKey": " demo-source ",
		"name": "Demo Source",
		"version": "1.0.0",
		"description": "bad",
		"runtime": { "type": "node" },
		"fetchPolicy": { "type": "fixed_interval", "minutes": 15 },
		"entrypoints": {
			"validate": { "command": ["node", "validate.mjs"] },
			"fetch": { "command": ["node", "fetch.mjs"] }
		},
		"workspaceConfigSchema": [
			{
				"key": " feedUrl ",
				"label": "Feed URL",
				"type": "url",
				"required": true
			}
		]
	}`)

	_, err := ParseManifest(raw)
	if !errors.Is(err, ErrInvalidPlugin) {
		t.Fatalf("expected invalid plugin error, got %v", err)
	}
}

func TestParseManifestRejectsBlankCommandEntries(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"schemaVersion": 2,
		"kind": "source",
		"pluginKey": "demo-source",
		"name": "Demo Source",
		"version": "1.0.0",
		"description": "bad",
		"runtime": { "type": "node" },
		"fetchPolicy": { "type": "fixed_interval", "minutes": 15 },
		"entrypoints": {
			"validate": { "command": ["   "] },
			"fetch": { "command": ["node", "fetch.mjs"] }
		},
		"workspaceConfigSchema": []
	}`)

	_, err := ParseManifest(raw)
	if !errors.Is(err, ErrInvalidPlugin) {
		t.Fatalf("expected invalid plugin error, got %v", err)
	}
}

func TestNormalizeConfigValuesRejectsFractionalNumbers(t *testing.T) {
	t.Parallel()

	_, _, errs := NormalizeConfigValues([]FieldSpec{
		{
			Key:      "repeat",
			Label:    "Repeat",
			Type:     FieldTypeNumber,
			Required: true,
		},
	}, map[string]any{
		"repeat": 1.5,
	}, false)

	if len(errs) != 1 || errs[0].Field != "repeat" {
		t.Fatalf("expected repeat field error, got %+v", errs)
	}
}
