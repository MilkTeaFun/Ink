package plugins

import (
	"errors"
	"testing"
)

func TestParseManifestRejectsSecretScheduleFields(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"schemaVersion": 1,
		"kind": "source",
		"pluginKey": "bad-schedule-secret",
		"name": "Bad Plugin",
		"version": "1.0.0",
		"description": "bad",
		"runtime": { "type": "node" },
		"entrypoints": {
			"validate": { "command": ["node", "validate.mjs"] },
			"fetch": { "command": ["node", "fetch.mjs"] }
		},
		"workspaceConfigSchema": [],
		"scheduleConfigSchema": [
			{
				"key": "token",
				"label": "Token",
				"type": "secret",
				"required": false
			}
		]
	}`)

	_, err := ParseManifest(raw)
	if !errors.Is(err, ErrInvalidPlugin) {
		t.Fatalf("expected invalid plugin error, got %v", err)
	}
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
		"schemaVersion": 1,
		"kind": "source",
		"pluginKey": " demo-source ",
		"name": "Demo Source",
		"version": "1.0.0",
		"description": "bad",
		"runtime": { "type": "node" },
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
		],
		"scheduleConfigSchema": []
	}`)

	_, err := ParseManifest(raw)
	if !errors.Is(err, ErrInvalidPlugin) {
		t.Fatalf("expected invalid plugin error, got %v", err)
	}
}

func TestParseManifestRejectsBlankCommandEntries(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"schemaVersion": 1,
		"kind": "source",
		"pluginKey": "demo-source",
		"name": "Demo Source",
		"version": "1.0.0",
		"description": "bad",
		"runtime": { "type": "node" },
		"entrypoints": {
			"validate": { "command": ["   "] },
			"fetch": { "command": ["node", "fetch.mjs"] }
		},
		"workspaceConfigSchema": [],
		"scheduleConfigSchema": []
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
