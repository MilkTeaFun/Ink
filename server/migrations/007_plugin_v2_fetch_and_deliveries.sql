alter table plugin_bindings
  add column if not exists next_fetch_at timestamptz null;

alter table plugin_bindings
  add column if not exists last_fetch_at timestamptz null;

alter table plugin_bindings
  add column if not exists fetch_lease_until timestamptz null;

alter table plugin_bindings
  add column if not exists last_fetch_error text null;

create index if not exists plugin_bindings_fetch_due_idx
  on plugin_bindings (enabled, status, next_fetch_at asc);

alter table print_schedules
  add column if not exists print_policy_json jsonb not null default '{"batchSize":1}'::jsonb;

create table if not exists print_schedule_deliveries (
  id text primary key,
  print_schedule_id text not null references print_schedules (id) on delete cascade,
  plugin_item_id text not null references plugin_items (id) on delete cascade,
  status text not null check (status in ('printed', 'failed')),
  attempt_count integer not null default 0,
  last_error text null,
  print_job_id text null references print_jobs (id) on delete set null,
  delivered_at timestamptz null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create unique index if not exists print_schedule_deliveries_schedule_item_unique
  on print_schedule_deliveries (print_schedule_id, plugin_item_id);

create index if not exists print_schedule_deliveries_status_updated_idx
  on print_schedule_deliveries (status, updated_at asc);
