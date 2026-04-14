-- Tray members may INSERT items but must not read other contributors' rows.
-- Visibility: tray owner sees all items; everyone sees rows they sourced (source_user_id).
-- Joined members use `tray contributed` / single-item flows for their own lines only.

drop policy if exists "items_select_visible" on public.items;

create policy "items_select_visible"
  on public.items for select to authenticated
  using (
    source_user_id = (select auth.uid())
    or public.tray_is_owner(tray_id, (select auth.uid()))
  );

comment on table public.tray_members is
  'Users who may add items on a tray. Item visibility is not granted by membership alone — see items policies.';
