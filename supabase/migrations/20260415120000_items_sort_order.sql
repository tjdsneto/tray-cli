-- Manual ordering of items within each tray (sort_order). New rows append to the end.

alter table public.items
  add column if not exists sort_order integer;

update public.items i
set sort_order = sub.rn
from (
  select
    id,
    row_number() over (
      partition by tray_id
      order by created_at asc, id asc
    ) as rn
  from public.items
) sub
where i.id = sub.id
  and (i.sort_order is distinct from sub.rn);

alter table public.items
  alter column sort_order set not null;

comment on column public.items.sort_order is 'Display order within the tray; lower values appear first in tray list.';

create index if not exists items_tray_id_sort_order_idx
  on public.items (tray_id, sort_order);

create or replace function public.items_assign_sort_order()
returns trigger
language plpgsql
security invoker
set search_path = public
as $$
begin
  if new.sort_order is null then
    new.sort_order := coalesce(
      (
        select max(i.sort_order)
        from public.items i
        where i.tray_id = new.tray_id
      ),
      0
    ) + 1;
  end if;
  return new;
end;
$$;

drop trigger if exists items_assign_sort_order_insert on public.items;

create trigger items_assign_sort_order_insert
  before insert on public.items
  for each row
  execute function public.items_assign_sort_order();
