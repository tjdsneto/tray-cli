-- Tray CLI core schema: trays, tray_members, items + RLS + join_tray(invite_token)

-- ---------------------------------------------------------------------------
-- Tables
-- ---------------------------------------------------------------------------

create table public.trays (
  id uuid primary key default gen_random_uuid(),
  owner_id uuid not null references auth.users (id) on delete cascade,
  name text not null,
  invite_token text,
  created_at timestamptz not null default now(),
  constraint trays_owner_name_unique unique (owner_id, name)
);

comment on table public.trays is 'A tray (shared inbox); owner_id is auth.users id.';

create unique index trays_invite_token_unique
  on public.trays (invite_token)
  where invite_token is not null;

create index trays_owner_id_idx on public.trays (owner_id);

create table public.tray_members (
  id uuid primary key default gen_random_uuid(),
  tray_id uuid not null references public.trays (id) on delete cascade,
  user_id uuid not null references auth.users (id) on delete cascade,
  joined_at timestamptz not null default now(),
  invited_via text,
  constraint tray_members_tray_user_unique unique (tray_id, user_id)
);

comment on table public.tray_members is 'Users who may add/read items on a tray (not necessarily the owner).';

create index tray_members_user_id_idx on public.tray_members (user_id);
create index tray_members_tray_id_idx on public.tray_members (tray_id);

create table public.items (
  id uuid primary key default gen_random_uuid(),
  tray_id uuid not null references public.trays (id) on delete cascade,
  source_user_id uuid not null references auth.users (id) on delete restrict,
  title text not null,
  status text not null default 'pending',
  due_date date,
  snooze_until timestamptz,
  decline_reason text,
  completion_message text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint items_status_check check (
    status in (
      'pending',
      'accepted',
      'declined',
      'snoozed',
      'completed',
      'archived'
    )
  )
);

comment on table public.items is 'Line items on a tray; triage fields owner-controlled.';

create index items_tray_id_idx on public.items (tray_id);
create index items_tray_status_created_idx
  on public.items (tray_id, status, created_at);
create index items_source_user_id_idx on public.items (source_user_id);
create index items_updated_at_idx on public.items (updated_at);

-- ---------------------------------------------------------------------------
-- updated_at
-- ---------------------------------------------------------------------------

create or replace function public.set_items_updated_at()
returns trigger
language plpgsql
security invoker
set search_path = public
as $$
begin
  new.updated_at := now();
  return new;
end;
$$;

create trigger items_set_updated_at
  before update on public.items
  for each row
  execute function public.set_items_updated_at();

-- ---------------------------------------------------------------------------
-- Join by invite token (Model B); SECURITY DEFINER — validates token server-side
-- ---------------------------------------------------------------------------

create or replace function public.join_tray(p_invite_token text)
returns uuid
language plpgsql
security definer
set search_path = public
as $$
declare
  v_tray_id uuid;
begin
  if p_invite_token is null or length(trim(p_invite_token)) = 0 then
    raise exception 'invalid token';
  end if;

  select t.id
    into v_tray_id
  from public.trays t
  where t.invite_token = p_invite_token;

  if v_tray_id is null then
    raise exception 'invalid or expired invite';
  end if;

  insert into public.tray_members (tray_id, user_id, invited_via)
  values (v_tray_id, auth.uid(), 'token')
  on conflict (tray_id, user_id) do nothing;

  return v_tray_id;
end;
$$;

revoke all on function public.join_tray(text) from public;
grant execute on function public.join_tray(text) to authenticated;

comment on function public.join_tray(text) is
  'Join current user to tray by invite_token; idempotent on duplicate member.';

-- ---------------------------------------------------------------------------
-- Row level security
-- ---------------------------------------------------------------------------

alter table public.trays enable row level security;
alter table public.tray_members enable row level security;
alter table public.items enable row level security;

-- trays
create policy "trays_select_visible"
  on public.trays for select to authenticated
  using (
    owner_id = (select auth.uid())
    or exists (
      select 1
      from public.tray_members tm
      where tm.tray_id = trays.id
        and tm.user_id = (select auth.uid())
    )
  );

create policy "trays_insert_owner"
  on public.trays for insert to authenticated
  with check (owner_id = (select auth.uid()));

create policy "trays_update_owner"
  on public.trays for update to authenticated
  using (owner_id = (select auth.uid()))
  with check (owner_id = (select auth.uid()));

create policy "trays_delete_owner"
  on public.trays for delete to authenticated
  using (owner_id = (select auth.uid()));

-- tray_members
create policy "tray_members_select_owner_or_self"
  on public.tray_members for select to authenticated
  using (
    user_id = (select auth.uid())
    or exists (
      select 1
      from public.trays t
      where t.id = tray_members.tray_id
        and t.owner_id = (select auth.uid())
    )
  );

create policy "tray_members_insert_owner"
  on public.tray_members for insert to authenticated
  with check (
    exists (
      select 1
      from public.trays t
      where t.id = tray_id
        and t.owner_id = (select auth.uid())
    )
  );

create policy "tray_members_delete_owner_or_self"
  on public.tray_members for delete to authenticated
  using (
    user_id = (select auth.uid())
    or exists (
      select 1
      from public.trays t
      where t.id = tray_members.tray_id
        and t.owner_id = (select auth.uid())
    )
  );

-- items
create policy "items_select_visible"
  on public.items for select to authenticated
  using (
    source_user_id = (select auth.uid())
    or exists (
      select 1
      from public.trays t
      where t.id = items.tray_id
        and t.owner_id = (select auth.uid())
    )
    or exists (
      select 1
      from public.tray_members tm
      where tm.tray_id = items.tray_id
        and tm.user_id = (select auth.uid())
    )
  );

create policy "items_insert_owner_or_member"
  on public.items for insert to authenticated
  with check (
    source_user_id = (select auth.uid())
    and (
      exists (
        select 1
        from public.trays t
        where t.id = tray_id
          and t.owner_id = (select auth.uid())
      )
      or exists (
        select 1
        from public.tray_members tm
        where tm.tray_id = tray_id
          and tm.user_id = (select auth.uid())
      )
    )
  );

create policy "items_update_owner"
  on public.items for update to authenticated
  using (
    exists (
      select 1
      from public.trays t
      where t.id = items.tray_id
        and t.owner_id = (select auth.uid())
    )
  )
  with check (
    exists (
      select 1
      from public.trays t
      where t.id = items.tray_id
        and t.owner_id = (select auth.uid())
    )
  );

create policy "items_delete_owner_or_pending_withdraw"
  on public.items for delete to authenticated
  using (
    exists (
      select 1
      from public.trays t
      where t.id = items.tray_id
        and t.owner_id = (select auth.uid())
    )
    or (
      source_user_id = (select auth.uid())
      and status = 'pending'
    )
  );

-- ---------------------------------------------------------------------------
-- Grants (API: authenticated JWT)
-- ---------------------------------------------------------------------------

grant usage on schema public to authenticated;

grant select, insert, update, delete on public.trays to authenticated;
grant select, insert, update, delete on public.tray_members to authenticated;
grant select, insert, update, delete on public.items to authenticated;
