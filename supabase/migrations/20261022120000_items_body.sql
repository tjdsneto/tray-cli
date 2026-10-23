-- Optional large text payload on items (create-time); title remains the short headline.

alter table public.items
  add column if not exists body text;

alter table public.items
  drop constraint if exists items_body_char_length_check;

alter table public.items
  add constraint items_body_char_length_check
  check (body is null or char_length(body) <= 100000);

comment on column public.items.body is
  'Optional large text payload (create-time); title remains the short headline.';
