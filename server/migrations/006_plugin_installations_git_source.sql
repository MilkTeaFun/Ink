alter table plugin_installations
  drop constraint if exists plugin_installations_source_type_check;

alter table plugin_installations
  add constraint plugin_installations_source_type_check
  check (source_type in ('upload', 'git'));

alter table plugin_installations
  add column if not exists repo_url text not null default '';

alter table plugin_installations
  add column if not exists repo_ref text not null default '';

alter table plugin_installations
  add column if not exists repo_commit_sha text not null default '';

alter table plugin_installations
  add column if not exists repo_subdir text not null default '';
