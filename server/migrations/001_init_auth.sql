create table if not exists users (
  id text primary key,
  email text not null,
  password_hash text not null,
  display_name text not null,
  status text not null check (status in ('active', 'disabled')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  last_login_at timestamptz null
);

create unique index if not exists users_email_unique on users (email);

create table if not exists auth_sessions (
  id text primary key,
  family_id text not null,
  user_id text not null references users (id) on delete cascade,
  refresh_token_hash text not null unique,
  client_type text not null,
  user_agent text null,
  ip_address text null,
  expires_at timestamptz not null,
  rotated_at timestamptz null,
  revoked_at timestamptz null,
  created_at timestamptz not null,
  last_used_at timestamptz not null
);

create index if not exists auth_sessions_family_idx on auth_sessions (family_id);
create index if not exists auth_sessions_user_idx on auth_sessions (user_id);

create table if not exists auth_audit_logs (
  id text primary key,
  user_id text null references users (id) on delete set null,
  event_type text not null,
  client_type text not null,
  ip_address text null,
  user_agent text null,
  detail jsonb null,
  created_at timestamptz not null
);

create index if not exists auth_audit_logs_user_idx on auth_audit_logs (user_id, created_at desc);
