-- Apply status timestamps when rows are inserted with a non-pending terminal status
-- (e.g. tray owner adds to their own tray with status accepted).

create or replace function public.items_apply_status_timestamps()
returns trigger
language plpgsql
security invoker
set search_path = public
as $$
begin
  if tg_op = 'INSERT' then
    if new.status = 'accepted' then
      new.accepted_at := coalesce(new.accepted_at, now());
    end if;
    if new.status = 'declined' then
      new.declined_at := coalesce(new.declined_at, now());
    end if;
    if new.status = 'completed' then
      new.completed_at := coalesce(new.completed_at, now());
    end if;
    if new.status = 'archived' then
      new.archived_at := coalesce(new.archived_at, now());
    end if;
    if new.status = 'snoozed' then
      new.snoozed_at := now();
    end if;
    return new;
  end if;

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

create trigger items_apply_status_timestamps_insert
  before insert on public.items
  for each row
  execute function public.items_apply_status_timestamps();
