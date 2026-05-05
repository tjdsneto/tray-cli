/** Build POST /rest/v1/items body — mirrors Go `newAddItemRequest`. */
export function buildAddItemBody(
  userId: string,
  trayId: string,
  title: string,
  trayOwnerId: string,
  dueDate?: string | null,
): Record<string, unknown> {
  const uid = userId.trim();
  const tid = trayId.trim();
  const tit = title.trim();
  const owner = trayOwnerId.trim();
  if (!uid || !tid || !tit || !owner) {
    throw new Error("missing user, tray, title, or tray owner");
  }
  const status = uid.toLowerCase() === owner.toLowerCase() ? "accepted" : "pending";
  const body: Record<string, unknown> = {
    tray_id: tid,
    source_user_id: uid,
    title: tit,
    status,
  };
  if (dueDate?.trim()) body.due_date = dueDate.trim();
  return body;
}

/** Build PATCH body for status-only triage. */
export function buildStatusPatch(status: string, extra?: Record<string, unknown>): Record<string, unknown> {
  const s = status.trim();
  if (!s) throw new Error("empty status");
  return extra && Object.keys(extra).length > 0 ? { status: s, ...extra } : { status: s };
}
