-- Break RLS recursion between public.trays and public.tray_members:
-- trays_select_visible queries tray_members; tray_members policies queried trays,
-- re-entering trays policies (PostgreSQL error 42P17).

create or replace function public.tray_is_owner(p_tray_id uuid, p_user_id uuid)
returns boolean
language sql
security definer
set search_path = public
stable
as $$
  select exists (
    select 1
    from public.trays t
    where t.id = p_tray_id
      and t.owner_id = p_user_id
  );
$$;

comment on function public.tray_is_owner(uuid, uuid) is
  'RLS helper: true if p_user_id owns the tray. SECURITY DEFINER avoids policy recursion.';

revoke all on function public.tray_is_owner(uuid, uuid) from public;
grant execute on function public.tray_is_owner(uuid, uuid) to authenticated;

-- tray_members: replace subqueries on trays with tray_is_owner
drop policy if exists "tray_members_select_owner_or_self" on public.tray_members;
create policy "tray_members_select_owner_or_self"
  on public.tray_members for select to authenticated
  using (
    user_id = (select auth.uid())
    or public.tray_is_owner(tray_id, (select auth.uid()))
  );

drop policy if exists "tray_members_insert_owner" on public.tray_members;
create policy "tray_members_insert_owner"
  on public.tray_members for insert to authenticated
  with check (
    public.tray_is_owner(tray_id, (select auth.uid()))
  );

drop policy if exists "tray_members_delete_owner_or_self" on public.tray_members;
create policy "tray_members_delete_owner_or_self"
  on public.tray_members for delete to authenticated
  using (
    user_id = (select auth.uid())
    or public.tray_is_owner(tray_id, (select auth.uid()))
  );

-- items: replace owner checks that selected from trays
drop policy if exists "items_select_visible" on public.items;
create policy "items_select_visible"
  on public.items for select to authenticated
  using (
    source_user_id = (select auth.uid())
    or public.tray_is_owner(tray_id, (select auth.uid()))
    or exists (
      select 1
      from public.tray_members tm
      where tm.tray_id = items.tray_id
        and tm.user_id = (select auth.uid())
    )
  );

drop policy if exists "items_insert_owner_or_member" on public.items;
create policy "items_insert_owner_or_member"
  on public.items for insert to authenticated
  with check (
    source_user_id = (select auth.uid())
    and (
      public.tray_is_owner(tray_id, (select auth.uid()))
      or exists (
        select 1
        from public.tray_members tm
        where tm.tray_id = tray_id
          and tm.user_id = (select auth.uid())
      )
    )
  );

drop policy if exists "items_update_owner" on public.items;
create policy "items_update_owner"
  on public.items for update to authenticated
  using (
    public.tray_is_owner(items.tray_id, (select auth.uid()))
  )
  with check (
    public.tray_is_owner(items.tray_id, (select auth.uid()))
  );

drop policy if exists "items_delete_owner_or_pending_withdraw" on public.items;
create policy "items_delete_owner_or_pending_withdraw"
  on public.items for delete to authenticated
  using (
    public.tray_is_owner(items.tray_id, (select auth.uid()))
    or (
      source_user_id = (select auth.uid())
      and status = 'pending'
    )
  );
