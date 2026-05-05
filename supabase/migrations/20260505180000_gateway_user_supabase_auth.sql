-- Persist Supabase Auth session for gateway API calls (PostgREST as authenticated user; RLS applies).

create table if not exists public.gateway_user_supabase_auth (
  user_id uuid primary key references auth.users (id) on delete cascade,
  access_token text not null,
  refresh_token text not null,
  access_expires_at timestamptz not null,
  updated_at timestamptz not null default now()
);

create index if not exists gateway_user_supabase_auth_expires_idx
  on public.gateway_user_supabase_auth (access_expires_at);

comment on table public.gateway_user_supabase_auth is
  'Supabase Auth tokens for ChatGPT gateway PostgREST calls; service role only.';

alter table public.gateway_user_supabase_auth enable row level security;

revoke all on public.gateway_user_supabase_auth from anon, authenticated;
