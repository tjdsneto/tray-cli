-- Gateway OAuth bootstrap tables (server-side token/code state for ChatGPT integration).

create table if not exists public.gateway_oauth_codes (
  id uuid primary key default gen_random_uuid(),
  code_hash text not null unique,
  user_id uuid not null references auth.users (id) on delete cascade,
  client_id text not null,
  redirect_uri text not null,
  scope text not null default '',
  expires_at timestamptz not null,
  consumed_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists gateway_oauth_codes_user_id_idx on public.gateway_oauth_codes (user_id);
create index if not exists gateway_oauth_codes_expires_at_idx on public.gateway_oauth_codes (expires_at);

create table if not exists public.gateway_refresh_tokens (
  id uuid primary key default gen_random_uuid(),
  token_hash text not null unique,
  user_id uuid not null references auth.users (id) on delete cascade,
  client_id text not null,
  scope text not null default '',
  expires_at timestamptz not null,
  revoked_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists gateway_refresh_tokens_user_id_idx on public.gateway_refresh_tokens (user_id);
create index if not exists gateway_refresh_tokens_expires_at_idx on public.gateway_refresh_tokens (expires_at);

alter table public.gateway_oauth_codes enable row level security;
alter table public.gateway_refresh_tokens enable row level security;

revoke all on public.gateway_oauth_codes from anon, authenticated;
revoke all on public.gateway_refresh_tokens from anon, authenticated;
