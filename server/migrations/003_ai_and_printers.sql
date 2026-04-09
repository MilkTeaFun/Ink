create table if not exists ai_provider_settings (
  setting_key text primary key check (setting_key = 'system'),
  provider_name text not null,
  provider_type text not null,
  base_url text not null,
  model text not null,
  api_key_ciphertext bytea not null,
  api_key_nonce bytea not null,
  updated_by text null references users (id) on delete set null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists printer_bindings (
  id text primary key,
  user_id text not null references users (id) on delete cascade,
  name text not null,
  note text not null default '',
  device_identifier text not null,
  provider_user_id integer not null,
  status text not null check (status in ('connected', 'pending', 'offline')),
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create unique index if not exists printer_bindings_user_device_unique
  on printer_bindings (user_id, device_identifier);

create index if not exists printer_bindings_user_updated_idx
  on printer_bindings (user_id, updated_at desc);

create table if not exists print_jobs (
  id text primary key,
  user_id text not null references users (id) on delete cascade,
  printer_binding_id text not null references printer_bindings (id) on delete cascade,
  title text not null,
  source text not null,
  content text not null,
  status text not null check (status in ('pending', 'queued', 'completed', 'failed', 'cancelled')),
  provider_print_content_id integer null,
  provider_smart_guid text null,
  error_message text null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create index if not exists print_jobs_user_updated_idx
  on print_jobs (user_id, updated_at desc);
