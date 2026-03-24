-- Public profile row per auth user (email + optional name) for CLI display on items.

create table public.profiles (
  id uuid primary key references auth.users (id) on delete cascade,
  email text,
  full_name text,
  updated_at timestamptz not null default now()
);

comment on table public.profiles is 'Mirror of auth user email/name for PostgREST; synced by trigger.';

alter table public.profiles enable row level security;

-- Read profiles for users who appear on items on trays you can access (owner or member).
create policy "profiles_select_visible"
  on public.profiles for select
  to authenticated
  using (
    id = (select auth.uid())
    or exists (
      select 1
      from public.items i
      join public.trays t on t.id = i.tray_id
      where i.source_user_id = profiles.id
        and (
          t.owner_id = (select auth.uid())
          or exists (
            select 1 from public.tray_members tm
            where tm.tray_id = t.id and tm.user_id = (select auth.uid())
          )
        )
    )
  );

grant select on table public.profiles to authenticated;

-- Sync from auth.users (insert + update email/metadata).
create or replace function public.sync_profile_from_auth_user()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
  insert into public.profiles (id, email, full_name, updated_at)
  values (
    new.id,
    new.email,
    nullif(trim(coalesce(new.raw_user_meta_data->>'full_name', new.raw_user_meta_data->>'name', '')), ''),
    now()
  )
  on conflict (id) do update
    set email = excluded.email,
        full_name = excluded.full_name,
        updated_at = now();
  return new;
end;
$$;

drop trigger if exists on_auth_user_sync_profile on auth.users;
create trigger on_auth_user_sync_profile
  after insert or update of email, raw_user_meta_data on auth.users
  for each row
  execute function public.sync_profile_from_auth_user();

-- Backfill existing users (migration runs as superuser).
insert into public.profiles (id, email, full_name, updated_at)
select
  u.id,
  u.email,
  nullif(trim(coalesce(u.raw_user_meta_data->>'full_name', u.raw_user_meta_data->>'name', '')), ''),
  now()
from auth.users u
on conflict (id) do update
  set email = excluded.email,
      full_name = excluded.full_name,
      updated_at = now();
