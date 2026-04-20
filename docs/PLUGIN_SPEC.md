# Ink Plugin Spec v2

This document defines the only supported plugin contract for Ink source plugins.

- `schemaVersion` must be `2`
- Plugins own data collection cadence through `fetchPolicy`
- Print schedules only decide when already-collected items are printed
- `scheduleConfig` and `scheduleConfigSchema` do not exist in v2

## Mental Model

Ink now splits plugin work into two independent loops:

1. Binding fetch loop
   - Every enabled binding is fetched automatically on the cadence declared by the plugin manifest.
   - Fetching stores deduplicated items in `plugin_items`.
   - Fetching never creates print jobs.

2. Print schedule loop
   - Each schedule selects already-fetched items for its binding.
   - Delivery tracking is per schedule in `print_schedule_deliveries`.
   - Multiple schedules on the same binding can print the same collected item independently.

## Manifest

### Required fields

```json
{
  "schemaVersion": 2,
  "kind": "source",
  "pluginKey": "example-source",
  "name": "Example Source",
  "version": "1.0.0",
  "description": "Fetches printable content for Ink.",
  "runtime": {
    "type": "node"
  },
  "fetchPolicy": {
    "type": "fixed_interval",
    "minutes": 15
  },
  "entrypoints": {
    "validate": {
      "command": ["node", "validate.mjs"]
    },
    "fetch": {
      "command": ["node", "fetch.mjs"]
    }
  },
  "workspaceConfigSchema": [
    {
      "key": "feedUrl",
      "label": "Feed URL",
      "type": "url",
      "required": true
    }
  ]
}
```

### Field rules

- `schemaVersion`: must be `2`
- `kind`: must be `source`
- `pluginKey`: lowercase letters, digits, and dashes
- `runtime.type`: `node` or `python`
- `fetchPolicy.type`: must be `fixed_interval`
- `fetchPolicy.minutes`: positive integer
- `workspaceConfigSchema`: binding-level config only

### Supported config field types

- `text`
- `textarea`
- `url`
- `number`
- `select`
- `checkbox`
- `secret`

`secret` fields are allowed only in `workspaceConfigSchema`. Their values are encrypted in binding storage and are passed to the plugin through the `secrets` object.

## Entrypoints

### `validate`

`validate` checks whether a workspace binding can be enabled.

Input:

```json
{
  "workspaceConfig": {
    "feedUrl": "https://example.com/feed"
  },
  "secrets": {
    "apiToken": "secret"
  }
}
```

Output:

```json
{
  "valid": true,
  "errors": []
}
```

Validation failures should use field-level errors when possible:

```json
{
  "valid": false,
  "errors": [
    {
      "field": "feedUrl",
      "message": "Feed URL is required"
    }
  ]
}
```

### `fetch`

`fetch` collects new content for one binding.

Input:

```json
{
  "workspaceConfig": {
    "feedUrl": "https://example.com/feed"
  },
  "secrets": {
    "apiToken": "secret"
  },
  "cursor": "opaque-cursor-from-last-run",
  "trigger": {
    "kind": "automatic",
    "scheduledFor": "2026-04-20T12:00:00Z",
    "triggeredAt": "2026-04-20T12:00:03Z",
    "timezone": "UTC"
  }
}
```

Notes:

- `cursor` is optional and is whatever the plugin returned last time
- `trigger.kind` is `automatic` for background fetches and `manual` for `POST /api/v1/plugins/{installationID}/run`
- `scheduleConfig` is never provided in v2

Output:

```json
{
  "items": [
    {
      "externalId": "item-123",
      "title": "Daily Digest",
      "sourceLabel": "Example Source",
      "publishedAt": "2026-04-20T11:55:00Z",
      "blocks": [
        { "type": "heading", "level": 1, "text": "Daily Digest" },
        { "type": "paragraph", "text": "Hello from Ink." }
      ]
    }
  ],
  "cursor": "next-opaque-cursor"
}
```

`externalId` is the dedupe key together with `plugin_binding_id`.

## Content Blocks

Supported block types:

- `heading`
- `paragraph`
- `image`
- `link`
- `divider`

Plugins should emit only simple, already-sanitized printable content.

## Binding Fetch State

Each binding tracks fetch execution separately from print schedules.

- `next_fetch_at`
- `last_fetch_at`
- `fetch_lease_until`
- `last_fetch_error`

Behavior:

- Enabling a binding schedules immediate fetch
- Disabling a binding clears future automatic fetches
- Successful fetch updates cursor and `next_fetch_at`
- Failed fetch updates `last_fetch_error` and schedules the next retry

## Collected Items And Deliveries

Fetched items are stored once per binding:

- `plugin_items` is deduplicated on `(plugin_binding_id, external_id)`

Printing is tracked per schedule:

- `print_schedule_deliveries` is deduplicated on `(print_schedule_id, plugin_item_id)`

Delivery behavior:

- Schedule ticks pick the oldest fetched items for the binding that do not yet have a delivery row for that schedule
- Failed print attempts are retried through the delivery row
- A successful delivery for schedule A does not block schedule B

## Print Schedules

Print schedules are platform-owned. Plugins do not define schedule-specific fields.

Schedule payloads use:

```json
{
  "title": "Morning Digest",
  "pluginInstallationId": "plugin-installation-1",
  "frequencyType": "daily",
  "timezone": "Asia/Shanghai",
  "hour": 9,
  "minute": 30,
  "weekdays": [],
  "printPolicy": {
    "batchSize": 1
  },
  "deviceId": "device-1",
  "enabled": true
}
```

`printPolicy` currently supports:

- `batchSize`: positive integer, defaults to `1`

Delivery order is fixed to oldest-first.

## Manual Operations

- `POST /api/v1/plugins/{installationID}/run`
  - Manual fetch only
  - Fetches and ingests items
  - Does not print

- `POST /api/v1/print-schedules/{scheduleID}/run`
  - Manual print tick only
  - Prints already-collected items for that schedule
  - Does not fetch

## Example Node Fetch Entrypoint

```js
import process from "node:process";

async function readJSON() {
  const chunks = [];
  for await (const chunk of process.stdin) chunks.push(chunk);
  const raw = Buffer.concat(chunks).toString("utf8").trim();
  return raw ? JSON.parse(raw) : {};
}

const payload = await readJSON();
const sourceName = String(payload.workspaceConfig?.sourceName ?? "Example Source").trim();

process.stdout.write(
  JSON.stringify({
    items: [
      {
        externalId: `example-${payload.trigger?.triggeredAt ?? "default"}`,
        title: `${sourceName} Digest`,
        sourceLabel: sourceName,
        blocks: [
          { type: "heading", level: 1, text: `${sourceName} Digest` },
          { type: "paragraph", text: "Hello from Ink." }
        ]
      }
    ],
    cursor: payload.trigger?.triggeredAt ?? null
  })
);
```

## Compatibility

v2 is a breaking contract.

- `schemaVersion: 1` is obsolete
- `scheduleConfigSchema` is obsolete
- `fetch(scheduleConfig)` is obsolete
- New plugins must target `schemaVersion: 2`
