const ITEM_SELECT =
  "id,tray_id,sort_order,source_user_id,title,status,due_date,snooze_until,decline_reason,completion_message,accepted_at,declined_at,completed_at,archived_at,snoozed_at,created_at,updated_at";

/** GET tray owner before insert (same as CLI add flow). */
export function trayOwnerSelectPath(trayId: string): string {
  const u = new URLSearchParams();
  u.set("select", "owner_id");
  u.set("id", `eq.${trayId.trim()}`);
  u.set("limit", "1");
  return `/rest/v1/trays?${u.toString()}`;
}

export function itemsInsertPath(): string {
  const u = new URLSearchParams();
  u.set("select", ITEM_SELECT);
  return `/rest/v1/items?${u.toString()}`;
}

export function itemsPatchPath(itemId: string): string {
  const u = new URLSearchParams();
  u.set("id", `eq.${itemId.trim()}`);
  return `/rest/v1/items?${u.toString()}`;
}

/** GET /rest/v1/trays — same as CLI `tray ls`: trays you own (`owner_id` filter). */
export function traysListOwnedPath(userId: string): string {
  const u = new URLSearchParams();
  u.set("select", "id,owner_id,name,invite_token,created_at,items(count)");
  u.set("owner_id", `eq.${userId.trim()}`);
  u.set("order", "name.asc");
  return `/rest/v1/trays?${u.toString()}`;
}

/** GET /rest/v1/tray_members — same source as CLI `tray remote ls` / ListJoined (filter out own trays in handler). */
export function trayMembersJoinedPath(userId: string): string {
  const u = new URLSearchParams();
  u.set("user_id", `eq.${userId.trim()}`);
  u.set("select", "joined_at,trays(id,owner_id,name,invite_token,created_at,items(count))");
  u.set("order", "joined_at.asc");
  return `/rest/v1/tray_members?${u.toString()}`;
}

export type ItemsListQuery = {
  trayId: string;
  status?: string;
  limit: number;
};

export function itemsListPath(q: ItemsListQuery): string {
  const u = new URLSearchParams();
  u.set("select", ITEM_SELECT);
  u.set("tray_id", `eq.${q.trayId}`);
  if (q.status?.trim()) u.set("status", `eq.${q.status.trim()}`);
  u.set("order", "sort_order.asc,created_at.asc");
  u.set("limit", String(q.limit));
  return `/rest/v1/items?${u.toString()}`;
}

export type ItemsAllVisibleQuery = {
  status?: string;
  trayId?: string;
  limit: number;
};

/** GET /rest/v1/items — no tray filter: RLS returns every item you may see (across trays). */
export function itemsListAllVisiblePath(q: ItemsAllVisibleQuery): string {
  const u = new URLSearchParams();
  u.set("select", ITEM_SELECT);
  if (q.trayId?.trim()) u.set("tray_id", `eq.${q.trayId.trim()}`);
  if (q.status?.trim()) u.set("status", `eq.${q.status.trim()}`);
  u.set("order", "updated_at.desc,created_at.desc");
  u.set("limit", String(q.limit));
  return `/rest/v1/items?${u.toString()}`;
}

/** Same PostgREST shape as CLI `tray contributed` before filtering out own trays in the handler. */
export type ItemsContributedQuery = {
  userId: string;
  status?: string;
  limit: number;
};

export function itemsContributedPath(q: ItemsContributedQuery): string {
  const u = new URLSearchParams();
  u.set("select", `${ITEM_SELECT},trays(owner_id)`);
  u.set("source_user_id", `eq.${q.userId.trim()}`);
  if (q.status?.trim()) u.set("status", `eq.${q.status.trim()}`);
  u.set("order", "sort_order.asc,created_at.asc");
  u.set("limit", String(q.limit));
  return `/rest/v1/items?${u.toString()}`;
}
