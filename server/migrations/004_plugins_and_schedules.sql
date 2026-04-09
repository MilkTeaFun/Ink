create table if not exists plugin_installations (
  id text primary key,
  plugin_key text not null unique,
  source_type text not null check (source_type in ('upload')),
  display_name text not null,
  version text not null,
  runtime_type text not null check (runtime_type in ('node', 'python')),
  manifest_json jsonb not null,
  current_path text not null default '',
  status text not null check (status in ('installing', 'ready', 'failed', 'disabled')),
  last_error text null,
  installed_by text null references users (id) on delete set null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create index if not exists plugin_installations_status_updated_idx
  on plugin_installations (status, updated_at desc);

create table if not exists plugin_bindings (
  id text primary key,
  plugin_installation_id text not null references plugin_installations (id) on delete cascade,
  user_id text not null references users (id) on delete cascade,
  enabled boolean not null default false,
  config_json jsonb not null default '{}'::jsonb,
  secret_ciphertext bytea null,
  secret_nonce bytea null,
  status text not null check (status in ('connected', 'disconnected', 'error')),
  last_validated_at timestamptz null,
  last_error text null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create unique index if not exists plugin_bindings_installation_user_unique
  on plugin_bindings (plugin_installation_id, user_id);

create index if not exists plugin_bindings_user_updated_idx
  on plugin_bindings (user_id, updated_at desc);

create table if not exists print_schedules (
  id text primary key,
  user_id text not null references users (id) on delete cascade,
  plugin_installation_id text not null references plugin_installations (id) on delete cascade,
  plugin_binding_id text not null references plugin_bindings (id) on delete cascade,
  title text not null,
  frequency_type text not null check (frequency_type in ('daily', 'weekly')),
  timezone text not null,
  hour integer not null check (hour >= 0 and hour <= 23),
  minute integer not null check (minute >= 0 and minute <= 59),
  weekdays jsonb not null default '[]'::jsonb,
  schedule_config_json jsonb not null default '{}'::jsonb,
  device_id text not null references printer_bindings (id) on delete cascade,
  enabled boolean not null default true,
  next_run_at timestamptz not null,
  last_run_at timestamptz null,
  lease_until timestamptz null,
  last_error text null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create index if not exists print_schedules_user_updated_idx
  on print_schedules (user_id, updated_at desc);

create index if not exists print_schedules_due_idx
  on print_schedules (enabled, next_run_at asc);
