alter table plugin_bindings
  add column if not exists cursor_json jsonb not null default 'null'::jsonb;

alter table plugin_bindings
  add column if not exists max_prints_per_run integer not null default 0;

alter table plugin_bindings
  add column if not exists max_prints_per_day integer not null default 0;

create table if not exists plugin_items (
  id text primary key,
  user_id text not null references users (id) on delete cascade,
  plugin_installation_id text not null references plugin_installations (id) on delete cascade,
  plugin_binding_id text not null references plugin_bindings (id) on delete cascade,
  device_id text null references printer_bindings (id) on delete set null,
  external_id text not null,
  title text not null,
  source_label text not null,
  published_at timestamptz null,
  blocks_json jsonb not null,
  status text not null check (status in ('pending', 'printed', 'invalid', 'failed')),
  attempt_count integer not null default 0,
  last_error text null,
  print_job_id text null references print_jobs (id) on delete set null,
  fetched_at timestamptz not null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create unique index if not exists plugin_items_dedup_idx
  on plugin_items (plugin_binding_id, external_id);

create index if not exists plugin_items_dispatch_idx
  on plugin_items (plugin_binding_id, status, created_at asc);

create index if not exists plugin_items_status_updated_idx
  on plugin_items (status, updated_at asc);

create index if not exists plugin_items_user_updated_idx
  on plugin_items (user_id, updated_at desc);
