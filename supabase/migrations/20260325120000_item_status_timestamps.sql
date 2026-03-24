-- When an item first enters accepted / declined / completed / archived / snoozed, record the time.
-- Client patches only send status (and optional messages); timestamps are applied in the database.

alter table public.items
  add column accepted_at timestamptz,
  add column declined_at timestamptz,
  add column completed_at timestamptz,
  add column archived_at timestamptz,
  add column snoozed_at timestamptz;

comment on column public.items.accepted_at is 'Set when status becomes accepted (first transition only).';
comment on column public.items.declined_at is 'Set when status becomes declined (first transition only).';
comment on column public.items.completed_at is 'Set when status becomes completed (first transition only).';
comment on column public.items.archived_at is 'Set when status becomes archived (first transition only).';
comment on column public.items.snoozed_at is 'Set each time status becomes snoozed.';

create or replace function public.items_apply_status_timestamps()
returns trigger
language plpgsql
security invoker
set search_path = public
as $$
begin
  if old.status is distinct from new.status then
    if new.status = 'accepted' and old.status is distinct from 'accepted' then
      new.accepted_at := coalesce(new.accepted_at, now());
    end if;
    if new.status = 'declined' and old.status is distinct from 'declined' then
      new.declined_at := coalesce(new.declined_at, now());
    end if;
    if new.status = 'completed' and old.status is distinct from 'completed' then
      new.completed_at := coalesce(new.completed_at, now());
    end if;
    if new.status = 'archived' and old.status is distinct from 'archived' then
      new.archived_at := coalesce(new.archived_at, now());
    end if;
    if new.status = 'snoozed' and old.status is distinct from 'snoozed' then
      new.snoozed_at := now();
    end if;
  end if;
  return new;
end;
$$;

create trigger items_apply_status_timestamps
  before update on public.items
  for each row
  execute function public.items_apply_status_timestamps();

-- Approximate historical times from last update (best effort for rows created before this migration).
update public.items set accepted_at = updated_at where status = 'accepted' and accepted_at is null;
update public.items set declined_at = updated_at where status = 'declined' and declined_at is null;
update public.items set completed_at = updated_at where status = 'completed' and completed_at is null;
update public.items set archived_at = updated_at where status = 'archived' and archived_at is null;
update public.items set snoozed_at = updated_at where status = 'snoozed' and snoozed_at is null;
