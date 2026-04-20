# Ink Plugin Specification (v1)

This document is the source of truth for the Ink plugin contract. It describes
what a plugin is, what Ink passes to a plugin on every invocation, and what a
plugin must return so Ink can ingest, deduplicate, and print the result.

The canonical types live in the Go source:

- `server/internal/plugins/plugins.go` — manifest, config, trigger, item, and
  fetch-output types
- `server/internal/plugins/manifest.go` — manifest validation and config
  normalization
- `server/internal/plugins/blocks.go` — block validation
- `server/internal/printer/blocks.go` — block-to-printer rendering

This spec targets `schemaVersion: 1`. Breaking changes will bump the version.

## 1. What a plugin is

A plugin is a directory containing:

1. An `ink-plugin.json` manifest at the directory root.
2. One or more executable files referenced by the manifest's `entrypoints`.
3. Optional runtime-specific dependency files (`package.json` + lockfile for
   Node, `pyproject.toml` + lockfile for Python).

Ink runs plugin entrypoints as ordinary subprocesses. Communication is
**JSON over stdio**: Ink writes a JSON request to the plugin's `stdin` and
reads a single JSON response from the plugin's `stdout`. A non-zero exit code
or invalid JSON output is treated as an execution failure.

The plugin is a **data source**. It converts an external system (RSS, Weibo,
Twitter, a webhook bridge, …) into a normalized list of printable items. It
does **not** choose which printer to use, enforce rate limits, or format for
the thermal head — Ink owns all of that.

## 2. Runtime

Ink currently supports two runtimes. The value of `runtime.type` in the
manifest controls which toolchain Ink uses at install time.

| runtime.type | Install command                     | Typical entrypoint command |
|--------------|-------------------------------------|----------------------------|
| `node`       | `pnpm install --frozen-lockfile`    | `["node", "fetch.mjs"]`    |
| `python`     | `uv sync --frozen`                  | `["python3", "fetch.py"]`  |

Other runtimes are rejected at manifest parse time.

Plugins must ship a lockfile that is compatible with the install command. The
install step runs inside Ink's `installTimeout` window (`PLUGIN_INSTALL_TIMEOUT`),
and failure to install marks the installation as `failed`.

Each entrypoint invocation runs with its own timeout (`PLUGIN_EXEC_TIMEOUT`,
20s by default). Anything that exceeds the timeout is killed and reported as
`ErrExecutionFailed`.

## 3. Manifest (`ink-plugin.json`)

The manifest is a JSON object with the following shape:

```json
{
  "schemaVersion": 1,
  "kind": "source",
  "pluginKey": "my-rss-source",
  "name": "My RSS Source",
  "version": "1.0.0",
  "description": "Fetches the latest items from a configured RSS feed.",
  "runtime": { "type": "node" },
  "entrypoints": {
    "validate": { "command": ["node", "validate.mjs"] },
    "fetch":    { "command": ["node", "fetch.mjs"] }
  },
  "workspaceConfigSchema": [ /* FieldSpec[] */ ],
  "scheduleConfigSchema":  [ /* FieldSpec[] */ ]
}
```

### Required fields

| Field            | Rules                                                                 |
|------------------|-----------------------------------------------------------------------|
| `schemaVersion`  | Must be `1`.                                                          |
| `kind`           | Must be `"source"`. Future kinds may be added additively.             |
| `pluginKey`      | Lowercase, digits, and dashes only (`^[a-z0-9][a-z0-9-]*$`). Must be unique across the Ink instance. |
| `name`           | Human-readable, non-empty.                                            |
| `version`        | Non-empty string. Semver is recommended but not enforced.             |
| `runtime.type`   | `"node"` or `"python"`.                                               |
| `entrypoints.validate.command` | Non-empty argv array; first element is the executable. |
| `entrypoints.fetch.command`    | Non-empty argv array; first element is the executable. |

`description` is optional and surfaced to admins in the plugin list.

### `workspaceConfigSchema` vs `scheduleConfigSchema`

Both schemas are arrays of `FieldSpec` objects. The distinction is about
**where the value is stored and how long it lives**:

- **`workspaceConfigSchema`** — Per-user, persistent binding config. Values are
  entered once when the user enables the plugin and persist until edited.
  **This is the only schema that may contain `secret` fields.** Secrets are
  stored encrypted (AES-GCM) on the binding record and decrypted only in-memory
  when an entrypoint runs.
- **`scheduleConfigSchema`** — Per-schedule config. Values are attached to each
  schedule that triggers the plugin, so a single binding can fan out into many
  schedules with different content. `secret` fields are **not allowed** here.

### `FieldSpec`

```ts
type FieldSpec = {
  key: string;            // unique within the schema; no surrounding whitespace
  label: string;          // non-empty human label
  type: "text" | "secret" | "textarea" | "url" | "number" | "select" | "checkbox";
  required?: boolean;
  description?: string;
  defaultValue?: unknown; // must round-trip through the type's normalizer
  options?: { label: string; value: string }[]; // required for type=select
};
```

Type-specific rules enforced by `NormalizeConfigValues`:

- `text` / `textarea` / `secret`: trimmed to non-empty string. `secret` is only
  allowed inside `workspaceConfigSchema`.
- `url`: must parse via `url.ParseRequestURI` with a non-empty scheme and host.
- `number`: accepts integers or integer-valued JSON numbers; fractional values
  are rejected.
- `checkbox`: accepts booleans or the strings `1/0/true/false/yes/no/on/off`.
- `select`: must exactly match one of `options[].value`.
- Unknown keys submitted by the user are rejected with
  `"包含未声明字段"` — the server will not silently drop them.

Default values are validated at manifest parse time: if `defaultValue` fails
normalization, the manifest is rejected.

## 4. Entrypoint I/O

All I/O is UTF-8 JSON. Ink writes exactly one JSON document to stdin (no
trailing newline is required but is allowed), then closes stdin. The plugin
must write exactly one JSON document to stdout and exit with code 0. stderr is
captured and included in error messages but is otherwise ignored.

### 4.1 `validate`

Called before Ink stores or enables a binding and when the admin hits the
"test binding" endpoint. Its only job is to tell Ink whether the current
`workspaceConfig` + `secrets` combination is viable.

**Request (stdin):**

```json
{
  "workspaceConfig": { "<key>": <value>, ... },
  "secrets":         { "<secretKey>": "<plaintext>", ... }
}
```

**Response (stdout):**

```json
{
  "valid": true,
  "errors": [ { "field": "<key>", "message": "<human message>" } ]
}
```

- `valid: true` with an empty `errors` array means "config is good".
- `valid: false` with an empty `errors` is replaced by a generic
  `"插件校验失败"` error so the UI always has something to display.
- `field` may be empty for non-field-scoped errors (e.g. authentication
  failure). The UI will render them at the top of the form.
- A non-zero exit or non-JSON stdout is surfaced as `ErrExecutionFailed` — do
  not rely on it to signal "invalid config". Always return structured errors.

### 4.2 `fetch`

Called by the scheduler on each tick that matches a bound schedule, and by the
manual `POST /api/v1/plugins/{installationID}/run` endpoint. It produces the
items Ink will ingest.

**Request (stdin):**

```json
{
  "workspaceConfig": { ... },
  "secrets":         { ... },
  "scheduleConfig":  { ... },
  "cursor":          "<opaque string or null>",
  "trigger": {
    "kind":         "schedule" | "manual",
    "scheduledFor": "2025-04-20T08:00:00Z",  // schedule kind only
    "triggeredAt":  "2025-04-20T08:00:03Z",
    "timezone":     "Asia/Shanghai"
  }
}
```

All timestamps are RFC 3339 strings. `scheduledFor` is omitted for manual
triggers. `cursor` is whatever string the plugin returned from its previous
successful fetch for this binding (or `null` on the first run).

**Response (stdout):**

```json
{
  "items": [
    {
      "externalId":  "2025-04-20-post-123",
      "title":       "Morning digest",
      "sourceLabel": "Example RSS",
      "publishedAt": "2025-04-20T07:55:00Z",
      "blocks":      [ /* see §5 */ ]
    }
  ],
  "cursor": "2025-04-20T08:00:00Z"
}
```

Field semantics:

- `items` is required; an empty array is a valid "nothing new" response.
- `externalId` is the **idempotency key**. The tuple
  `(plugin_binding_id, external_id)` has a unique index in the inbox, so Ink
  silently drops items whose external id it already ingested for the same
  binding. Pick something stable and per-source: post id, feed entry GUID,
  tweet id, etc. Do **not** put the current timestamp in it unless you
  actually want every run to produce a new item.
- `title` is required and will be used as the print job's title. Keep it
  short.
- `sourceLabel` is optional. If the plugin omits it or emits whitespace, Ink
  fills it in with the installation's `displayName`.
- `publishedAt` is optional; when present it must be RFC 3339. It is stored
  verbatim and may be surfaced in future UIs.
- `blocks` is required and must contain at least one valid block (see §5).
- `cursor` is optional. On success Ink persists it on the binding, overwriting
  any previous value, and will pass it back as `cursor` on the next call. It
  is opaque to Ink — use whatever serialized state (timestamp, page token,
  ETag, high-watermark id, …) your source needs. Returning `null` or omitting
  the field leaves the stored cursor unchanged on a failed run and clears it
  on a successful run.

## 5. Content blocks

Plugins emit structured `ContentBlock`s rather than raw strings. Ink's printer
renderer projects them to plain text for the thermal head; future UIs may
render them more richly without a plugin-side change.

Exactly 5 block types are supported. Unknown types are rejected at ingest
time and mark the item as `invalid`.

| type        | Required fields | Rendered as                                          |
|-------------|-----------------|------------------------------------------------------|
| `heading`   | `level` (1–3), `text` | `#`, `##`, or `###` prefix + text              |
| `paragraph` | `text`                | Plain line                                       |
| `image`     | `url` (http/https), optional `alt` | `[<alt or "图片">] <url>`           |
| `link`      | `url` (http/https), optional `text` | `<text>\n<url>` (or just `<url>`)  |
| `divider`   | —                     | 16-dash separator                                |

Validation rules enforced by `plugins.ValidateBlocks`:

- `heading.level` must be 1, 2, or 3.
- `heading.text` and `paragraph.text` must be non-empty after trimming.
- `image.url` and `link.url` must parse to an `http`/`https` URL with a host.
- Each item must contain at least one block.

Malformed blocks fail the entire item at ingest time; the dispatcher never
sees partial items.

## 6. Runtime environment

When Ink runs an entrypoint:

- **Working directory** is the installation root (the directory that contains
  `ink-plugin.json`).
- The command is the `command` array from the manifest, executed directly —
  no shell, no glob expansion, no PATH rewriting beyond what the host OS
  already does. If you need a shell, invoke it explicitly: `["bash", "-c",
  "..."]`.
- **stdin** is the JSON request; **stdout** must be the JSON response.
  **stderr** may be used for log lines; it is captured into error messages if
  the command fails.
- **Environment variables** are inherited from the Ink process. Do not rely on
  custom env vars for plugin configuration — use the manifest schemas and the
  `secrets` payload instead.
- Each invocation is ephemeral. Plugins must not assume any shared in-memory
  state between calls and should persist nothing to disk; use the `cursor`
  field to carry state forward.
- Network access is allowed. Outbound requests are subject only to host-level
  firewalling.

## 7. Installation

The only install source in this release is ZIP upload via
`POST /api/v1/admin/plugins/upload` (admin-only). The archive is unzipped into
a staging directory, the manifest is validated, runtime dependencies are
installed, and the directory is moved to
`<pluginRoot>/installations/<installationID>-<timestamp>/`.

Installation can fail for any of the following reasons:

- Missing or invalid `ink-plugin.json`.
- Unsupported `runtime.type`.
- Missing referenced entrypoint files.
- `pnpm install --frozen-lockfile` or `uv sync --frozen` exits non-zero.
- Install exceeds `installTimeout`.

A failed install is recorded with `status: "failed"` and a `lastError`
message so admins can diagnose from the UI.

Re-uploading an archive with the same `pluginKey` replaces the existing
installation in place; the old directory is atomically swapped out.

## 8. Runtime pipeline

```
Trigger (schedule | manual)
      ↓
PluginRuntime.fetch → Items[]
      ↓
Inbox   (plugin_items; dedup on (binding_id, external_id))
      ↓
Dispatcher (per-binding rate limits + retries)
      ↓
printer.CreatePrintJob
```

Key knobs plugins should be aware of — these are enforced by Ink, not by the
plugin, but they shape what behaviour makes sense:

- Per-run and per-day print caps are stored on the binding (`MaxPrintsPerRun`,
  `MaxPrintsPerDay`) with defaults of 20 and 50 respectively. Returning
  thousands of items in one fetch is safe — Ink will throttle the dispatcher
  — but the extra items will simply sit in the inbox until the cap frees up.
- Failed dispatches are retried by the `DispatchRunner` on a `DISPATCH_RETRY_INTERVAL`
  loop (default 1 min) up to `inbox.MaxDispatchAttempts = 3` attempts per item.
- Printed items are purged by the `InboxJanitor` after `INBOX_RETENTION`
  (default 30 days). Non-printed rows are never auto-purged.

Plugins are therefore free to be **stateless and greedy**: fetch everything
new since `cursor`, return it, update `cursor`, and let Ink handle the rest.

## 9. Worked example

A minimal Node plugin that echoes a configured message:

**`ink-plugin.json`**

```json
{
  "schemaVersion": 1,
  "kind": "source",
  "pluginKey": "hello-source",
  "name": "Hello Source",
  "version": "1.0.0",
  "description": "Prints a configurable greeting.",
  "runtime": { "type": "node" },
  "entrypoints": {
    "validate": { "command": ["node", "validate.mjs"] },
    "fetch":    { "command": ["node", "fetch.mjs"] }
  },
  "workspaceConfigSchema": [
    { "key": "sourceName", "label": "Source Name", "type": "text", "required": true }
  ],
  "scheduleConfigSchema": [
    { "key": "message", "label": "Message", "type": "textarea", "required": true }
  ]
}
```

**`validate.mjs`**

```js
import process from "node:process";

const raw = await new Promise((resolve) => {
  const chunks = [];
  process.stdin.on("data", (c) => chunks.push(c));
  process.stdin.on("end", () => resolve(Buffer.concat(chunks).toString("utf8")));
});
const payload = raw.trim() ? JSON.parse(raw) : {};
const errors = [];
if (!String(payload.workspaceConfig?.sourceName ?? "").trim()) {
  errors.push({ field: "sourceName", message: "sourceName is required" });
}
process.stdout.write(JSON.stringify({ valid: errors.length === 0, errors }));
```

**`fetch.mjs`**

```js
import process from "node:process";

const raw = await new Promise((resolve) => {
  const chunks = [];
  process.stdin.on("data", (c) => chunks.push(c));
  process.stdin.on("end", () => resolve(Buffer.concat(chunks).toString("utf8")));
});
const payload = raw.trim() ? JSON.parse(raw) : {};

const source = String(payload.workspaceConfig?.sourceName ?? "Hello").trim();
const message = String(payload.scheduleConfig?.message ?? "Hello").trim();
const triggeredAt = payload.trigger?.triggeredAt ?? new Date().toISOString();

process.stdout.write(JSON.stringify({
  items: [
    {
      externalId: `hello-${triggeredAt}`,
      title: `${source} Digest`,
      sourceLabel: source,
      blocks: [
        { type: "heading", level: 1, text: `${source} Digest` },
        { type: "paragraph", text: message },
      ],
    },
  ],
  cursor: triggeredAt,
}));
```

Canonical fixtures for both runtimes live under
`server/testdata/plugins/node-hello-plugin/` and
`server/testdata/plugins/python-hello-plugin/` and are exercised by the
server test suite.

## 10. Compatibility and versioning

- The manifest `schemaVersion` is `1`. Future breaking changes will bump the
  version and Ink will refuse to install older-or-newer manifests without
  explicit runtime support.
- New block types, new field types, and new entrypoints will be **additive**:
  existing `schemaVersion: 1` plugins continue to work, and plugins that
  opt into new features should also bump their own `version` so admins can
  track the upgrade.
- The `cursor` field is opaque to Ink and therefore free to change format
  between plugin versions, but plugins must be prepared to receive the cursor
  their previous version wrote. When changing cursor format, either:
  - version-prefix the cursor (e.g. `v2:<state>`) and handle old prefixes, or
  - treat unrecognized cursors as "start from the beginning".
