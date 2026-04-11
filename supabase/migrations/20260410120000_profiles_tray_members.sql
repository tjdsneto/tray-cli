-- Allow reading profiles for users in tray_members (owner + joiners), not only users who appear on items.

drop policy "profiles_select_visible" on public.profiles;

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
    or exists (
      select 1
      from public.tray_members tm
      join public.trays t on t.id = tm.tray_id
      where tm.user_id = profiles.id
        and (
          t.owner_id = (select auth.uid())
          or exists (
            select 1 from public.tray_members tm2
            where tm2.tray_id = t.id and tm2.user_id = (select auth.uid())
          )
        )
    )
  );
