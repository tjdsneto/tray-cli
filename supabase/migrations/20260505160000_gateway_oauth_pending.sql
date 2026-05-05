-- Pending ChatGPT OAuth sessions (PKCE verifier + params) until Supabase Auth returns.

create table if not exists public.gateway_oauth_pending (
  id uuid primary key default gen_random_uuid(),
  code_verifier text not null,
  client_id text not null,
  redirect_uri text not null,
  state text not null,
  scope text not null default '',
  expires_at timestamptz not null,
  created_at timestamptz not null default now()
);

create index if not exists gateway_oauth_pending_expires_at_idx on public.gateway_oauth_pending (expires_at);

alter table public.gateway_oauth_pending enable row level security;

revoke all on public.gateway_oauth_pending from anon, authenticated;
