-- Source AI-agent session id that created the item (GF-37-aligned; optional).

alter table public.items
  add column if not exists agent_session_id text;

comment on column public.items.agent_session_id is
  'Optional originating AI-agent session id (source). Target inbox is the agent-session:<id> tray name.';
