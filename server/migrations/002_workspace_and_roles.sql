alter table users
  add column if not exists role text not null default 'member'
  check (role in ('admin', 'member'));

update users
set role = 'admin'
where id = 'user_admin' or email = 'admin';

create table if not exists workspace_snapshots (
  user_id text primary key references users (id) on delete cascade,
  state jsonb not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists workspace_snapshots_updated_idx
  on workspace_snapshots (updated_at desc);
